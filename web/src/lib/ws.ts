import type { Snapshot } from './types';

export type ConnectionStatus = 'connecting' | 'live' | 'offline';

export interface RoomSocket {
  close: () => void;
  send: (data: unknown) => void;
}

export function connectRoom(
  code: string,
  client: 'guest' | 'tv' | 'phone',
  handlers: {
    onSnapshot: (snap: Snapshot) => void;
    onEnded: () => void;
    onStatus: (status: ConnectionStatus) => void;
  }
): RoomSocket {
  let ws: WebSocket | null = null;
  let closed = false;
  let retry = 0;
  let heartbeat: ReturnType<typeof setInterval> | null = null;
  let retryTimer: ReturnType<typeof setTimeout> | null = null;

  const clear = () => {
    if (heartbeat) clearInterval(heartbeat);
    heartbeat = null;
    if (retryTimer) clearTimeout(retryTimer);
    retryTimer = null;
  };

  const open = () => {
    if (closed) return;
    handlers.onStatus(retry === 0 ? 'connecting' : 'offline');
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    ws = new WebSocket(
      `${proto}://${location.host}/ws?room=${encodeURIComponent(code)}&client=${client}`
    );
    ws.onopen = () => {
      retry = 0;
      handlers.onStatus('live');
      heartbeat = setInterval(() => {
        if (ws?.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'heartbeat' }));
        }
      }, 20_000);
    };
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'snapshot') {
          const { type: _type, ...snap } = msg;
          handlers.onSnapshot(snap as Snapshot);
        } else if (msg.type === 'roomEnded') {
          handlers.onEnded();
        }
      } catch {
        // ignore malformed frames
      }
    };
    ws.onclose = (event) => {
      clear();
      if (event.code === 4001) {
        handlers.onEnded();
        return;
      }
      if (closed) return;
      handlers.onStatus('offline');
      const delay = Math.min(10_000, 500 * 2 ** retry) + Math.random() * 300;
      retry += 1;
      retryTimer = setTimeout(open, delay);
    };
    ws.onerror = () => {
      ws?.close();
    };
  };

  const onVisible = () => {
    if (document.visibilityState === 'visible' && (!ws || ws.readyState > 1)) {
      if (retryTimer) clearTimeout(retryTimer);
      retryTimer = null;
      open();
    }
  };
  document.addEventListener('visibilitychange', onVisible);
  open();

  return {
    close() {
      closed = true;
      clear();
      document.removeEventListener('visibilitychange', onVisible);
      ws?.close();
    },
    send(data: unknown) {
      if (ws?.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(data));
      }
    }
  };
}
