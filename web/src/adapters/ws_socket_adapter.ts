import { RECONNECT_BASE_MS, RECONNECT_MAX_MS } from '../domain/player_policy';
import type { Snapshot } from '../domain/models';
import type { ClientRole, RoomSocketHandlers, RoomSocketPort } from '../ports/socket_port';

export function connectRoom(
  code: string,
  client: ClientRole,
  handlers: RoomSocketHandlers
): RoomSocketPort {
  const socket = new SocketSession(code, client, handlers);
  document.addEventListener('visibilitychange', socket.onVisible);
  socket.open();
  return { close: () => socket.close(), send: (data) => socket.send(data) };
}

class SocketSession {
  private ws: WebSocket | null = null;
  private closed = false;
  private retry = 0;
  private heartbeat: ReturnType<typeof setInterval> | null = null;
  private retryTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(
    private readonly code: string,
    private readonly client: ClientRole,
    private readonly handlers: RoomSocketHandlers
  ) {}

  readonly onVisible = (): void => {
    if (document.visibilityState === 'visible' && (!this.ws || this.ws.readyState > 1)) {
      if (this.retryTimer) clearTimeout(this.retryTimer);
      this.retryTimer = null;
      this.open();
    }
  };

  open(): void {
    if (this.closed) return;
    this.handlers.onStatus(this.retry === 0 ? 'connecting' : 'offline');
    this.ws = new WebSocket(this.endpoint());
    this.ws.onopen = () => this.onOpen();
    this.ws.onmessage = (event) => this.onMessage(event);
    this.ws.onclose = (event) => this.onClose(event);
    this.ws.onerror = () => this.ws?.close();
  }

  send(data: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) this.ws.send(JSON.stringify(data));
  }

  close(): void {
    this.closed = true;
    this.clearTimers();
    document.removeEventListener('visibilitychange', this.onVisible);
    this.ws?.close();
  }

  private endpoint(): string {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    return `${proto}://${location.host}/ws?room=${encodeURIComponent(this.code)}&client=${this.client}`;
  }

  private onOpen(): void {
    this.retry = 0;
    this.handlers.onStatus('live');
    this.heartbeat = setInterval(() => this.send({ type: 'heartbeat' }), 20_000);
  }

  private onMessage(event: MessageEvent): void {
    let msg: { type?: string; [key: string]: unknown };
    try {
      msg = JSON.parse(event.data);
    } catch {
      return;
    }
    if (msg.type === 'snapshot') {
      const { type: _type, ...snap } = msg;
      this.handlers.onSnapshot(snap as unknown as Snapshot);
    } else if (msg.type === 'roomEnded') {
      this.handlers.onEnded();
    }
  }

  private onClose(event: CloseEvent): void {
    this.clearTimers();
    if (event.code === 4001) {
      this.handlers.onEnded();
      return;
    }
    if (this.closed) return;
    this.handlers.onStatus('offline');
    const delay = Math.min(RECONNECT_MAX_MS, RECONNECT_BASE_MS * 2 ** this.retry) + Math.random() * 300;
    this.retry += 1;
    this.retryTimer = setTimeout(() => this.open(), delay);
  }

  private clearTimers(): void {
    if (this.heartbeat) clearInterval(this.heartbeat);
    this.heartbeat = null;
    if (this.retryTimer) clearTimeout(this.retryTimer);
    this.retryTimer = null;
  }
}
