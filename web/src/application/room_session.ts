import { writable, type Writable } from 'svelte/store';
import type { Snapshot } from '../domain/models';
import { errorCodeOf, isRoomGone } from '../domain/errors';
import type { RoomApiPort } from '../ports/api_port';
import type { ClientRole, ConnectionStatus, SocketFactory } from '../ports/socket_port';
import { errorText, noticeText, type Translate } from './i18n';

export interface RoomSessionDeps {
  api: RoomApiPort;
  connectSocket: SocketFactory;
  code: () => string;
  client: ClientRole;
  translate: () => Translate;
  onEnded?: () => void;
}

export interface RoomSession {
  snap: Writable<Snapshot | null>;
  status: Writable<ConnectionStatus>;
  ended: Writable<boolean>;
  toast: Writable<string>;
  apply(snapshot: Snapshot): void;
  showToast(text: string): void;
  showError(err: unknown, fallback?: string): void;
  send(data: unknown): void;
  connect(): void;
  disconnect(): void;
}

export function createRoomSession(deps: RoomSessionDeps): RoomSession {
  const snap = writable<Snapshot | null>(null);
  const status = writable<ConnectionStatus>('connecting');
  const ended = writable(false);
  const toast = writable('');
  const session = new Session(deps, { snap, status, ended, toast });
  return {
    snap,
    status,
    ended,
    toast,
    apply: (snapshot) => session.apply(snapshot),
    showToast: (text) => session.showToast(text),
    showError: (err, fallback) => session.showError(err, fallback),
    send: (data) => session.send(data),
    connect: () => session.connect(),
    disconnect: () => session.disconnect()
  };
}

class Session {
  private socket: ReturnType<SocketFactory> | null = null;
  private seenFirst = false;
  private lastNoticeId = '';
  private toastTimer: ReturnType<typeof setTimeout> | null = null;
  private roomCheckTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(
    private readonly deps: RoomSessionDeps,
    private readonly stores: {
      snap: Writable<Snapshot | null>;
      status: Writable<ConnectionStatus>;
      ended: Writable<boolean>;
      toast: Writable<string>;
    }
  ) {}

  apply(snapshot: Snapshot): void {
    this.stores.snap.set(snapshot);
    const latest = snapshot.notices[0];
    if (!this.seenFirst) {
      this.seenFirst = true;
      this.lastNoticeId = latest?.id ?? '';
      return;
    }
    if (!latest || latest.id === this.lastNoticeId) return;
    this.lastNoticeId = latest.id;
    this.showToast(noticeText(this.deps.translate(), latest.code, latest.params));
  }

  showToast(text: string): void {
    if (!text) return;
    this.stores.toast.set(text);
    if (this.toastTimer) clearTimeout(this.toastTimer);
    this.toastTimer = setTimeout(() => this.stores.toast.set(''), 4000);
  }

  showError(err: unknown, fallback = 'generic'): void {
    this.showToast(errorText(this.deps.translate(), errorCodeOf(err, fallback)));
  }

  connect(): void {
    this.socket?.close();
    this.socket = this.deps.connectSocket(this.deps.code(), this.deps.client, {
      onStatus: (value) => this.onStatus(value),
      onSnapshot: (snapshot) => this.apply(snapshot),
      onEnded: () => this.end()
    });
  }

  disconnect(): void {
    this.socket?.close();
    this.socket = null;
    if (this.toastTimer) clearTimeout(this.toastTimer);
    if (this.roomCheckTimer) clearTimeout(this.roomCheckTimer);
  }

  send(data: unknown): void {
    this.socket?.send(data);
  }

  private onStatus(value: ConnectionStatus): void {
    this.stores.status.set(value);
    if (value === 'offline') {
      this.scheduleRoomCheck();
      return;
    }
    if (this.roomCheckTimer) clearTimeout(this.roomCheckTimer);
    this.roomCheckTimer = null;
  }

  private scheduleRoomCheck(): void {
    if (this.roomCheckTimer) return;
    this.roomCheckTimer = setTimeout(async () => {
      this.roomCheckTimer = null;
      try {
        await this.deps.api.getRoom(this.deps.code());
      } catch (err) {
        if (isRoomGone(err)) this.end();
      }
    }, 8000);
  }

  private end(): void {
    this.stores.ended.set(true);
    this.socket?.close();
    this.deps.onEnded?.();
  }
}
