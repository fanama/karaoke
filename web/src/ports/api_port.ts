import type {
  CreateRoomResponse,
  PlaybackAction,
  SearchResult,
  Snapshot,
  TVSession
} from '../domain/models';

export interface RoomApiPort {
  config(): Promise<{ maxGuests: number; maxQueue: number; languages: string[] }>;
  createRoom(): Promise<CreateRoomResponse>;
  resolveRoom(code: string): Promise<{ code: string }>;
  getRoom(code: string): Promise<Snapshot>;
  search(query: string): Promise<{ results: SearchResult[]; stale: boolean }>;
  addSong(
    code: string,
    payload: { videoId?: string; url?: string; nickname: string }
  ): Promise<Snapshot>;
  toggleVote(code: string, entryId: string): Promise<Snapshot>;
  startSkipVote(code: string): Promise<Snapshot>;
  castSkipVote(code: string): Promise<Snapshot>;
  cancelSkipVote(code: string): Promise<Snapshot>;
  unlockHost(code: string, pin: string): Promise<Snapshot>;
  playback(code: string, action: PlaybackAction): Promise<Snapshot>;
  removeSong(code: string, entryId: string): Promise<Snapshot>;
  reorder(code: string, entryIds: string[]): Promise<Snapshot>;
  clearQueue(code: string): Promise<Snapshot>;
  lockQueue(code: string, locked: boolean): Promise<Snapshot>;
  rotatePIN(code: string): Promise<{ pin: string }>;
  endRoom(code: string): Promise<void>;
  tvSession(): Promise<TVSession>;
}
