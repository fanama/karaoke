export type Status = 'idle' | 'countdown' | 'playing' | 'paused';

export interface Notice {
  id: string;
  code: string;
  params: Record<string, string>;
  createdAt: string;
}

export interface QueueItem {
  id: string;
  position: number;
  videoId: string;
  title: string;
  channel: string;
  thumbnailUrl: string;
  durationSeconds: number;
  nickname: string;
  score: number;
  votedByMe: boolean;
  mine: boolean;
}

export interface NowPlaying {
  entryId: string;
  videoId: string;
  title: string;
  channel: string;
  thumbnailUrl: string;
  nickname: string;
  durationSeconds: number;
  status: Status;
  elapsedSeconds: number;
  generation: number;
  countdownEndsAt: string | null;
  mine: boolean;
}

export interface SkipVote {
  active: boolean;
  yes: number;
  threshold: number;
  eligible: number;
  expiresAt: string | null;
  votedByMe: boolean;
  canVote: boolean;
  cooldownUntil: string | null;
  initiator?: string;
}

export interface RoomView {
  code: string;
  queueLocked: boolean;
  manualOrder: boolean;
  guestCount: number;
  guestCapacity: number;
  queueCapacity: number;
  tvStarted: boolean;
}

export interface YouView {
  deviceId: string;
  role: 'guest' | 'phone' | 'tv';
  hostUnlocked: boolean;
  canSkipVote: boolean;
  inQueue: boolean;
}

export interface Snapshot {
  seq: number;
  serverTime: string;
  room: RoomView;
  nowPlaying: NowPlaying | null;
  queue: QueueItem[];
  skipVote: SkipVote | null;
  notices: Notice[];
  you: YouView;
}

export interface SearchResult {
  videoId: string;
  title: string;
  channel: string;
  thumbnailUrl: string;
  durationSeconds: number;
}

export interface CreateRoomResponse {
  code: string;
  joinUrl: string;
  tvUrl: string;
  pin: string;
}

export interface TVSession {
  code: string;
  pin: string;
}

export interface AdminSession {
  enabled: boolean;
  authenticated: boolean;
  library: boolean;
  songs: number;
}

export interface QueueSelection {
  videoId?: string;
  url?: string;
  title?: string;
  channel?: string;
  thumbnailUrl?: string;
}

export type PlaybackAction = 'play' | 'pause' | 'next' | 'cancelCountdown';
