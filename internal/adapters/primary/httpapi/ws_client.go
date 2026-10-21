package httpapi

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type wsClient struct {
	code        string
	device      string
	role        string
	conn        *websocket.Conn
	send        chan []byte
	done        chan struct{}
	once        sync.Once
	closeCode   websocket.StatusCode
	closeReason string
}

func (c *wsClient) close() {
	c.once.Do(func() { close(c.done) })
}

func (c *wsClient) closeWith(code websocket.StatusCode, reason string) {
	c.closeCode = code
	c.closeReason = reason
}

// writePump writes queued messages and keeps the socket alive with pings.
func (c *wsClient) writePump() {
	ctx := context.Background()
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			c.finish()
			return
		case data := <-c.send:
			if !c.write(ctx, data) {
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				c.close()
				return
			}
		}
	}
}

func (c *wsClient) write(ctx context.Context, data []byte) bool {
	writeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	err := c.conn.Write(writeCtx, websocket.MessageText, data)
	cancel()
	if err != nil {
		c.close()
		return false
	}
	return true
}

func (c *wsClient) finish() {
	code := c.closeCode
	reason := c.closeReason
	if code == 0 {
		code = websocket.StatusNormalClosure
		reason = "bye"
	}
	_ = c.conn.Close(code, reason)
}
