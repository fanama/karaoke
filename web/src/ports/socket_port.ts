import type { Snapshot } from '../domain/models';

export type ConnectionStatus = 'connecting' | 'live' | 'offline';

export type ClientRole = 'guest' | 'tv' | 'phone';

export interface RoomSocketHandlers {
  onSnapshot: (snap: Snapshot) => void;
  onEnded: () => void;
  onStatus: (status: ConnectionStatus) => void;
}

export interface RoomSocketPort {
  close: () => void;
  send: (data: unknown) => void;
}

export type SocketFactory = (
  code: string,
  client: ClientRole,
  handlers: RoomSocketHandlers
) => RoomSocketPort;
