import type {
  CreateRoomResponse,
  SearchResult,
  Snapshot,
  TVSession
} from './types';

export class ApiError extends Error {
  code: string;
  status: number;
  retryAfter: number;

  constructor(code: string, message: string, status: number, retryAfter = 0) {
    super(message);
    this.code = code;
    this.status = status;
    this.retryAfter = retryAfter;
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  const res = await fetch(path, { ...init, headers, credentials: 'same-origin' });
  if (!res.ok) {
    let body: any = null;
    try {
      body = await res.json();
    } catch {
      // non-JSON error
    }
    throw new ApiError(
      body?.error?.code ?? 'internal_error',
      body?.error?.message ?? res.statusText,
      res.status,
      Number(body?.error?.retry_after ?? res.headers.get('Retry-After') ?? 0)
    );
  }
  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}

const post = (body?: unknown): RequestInit => ({
  method: 'POST',
  body: body === undefined ? undefined : JSON.stringify(body)
});

export const api = {
  config: () =>
    request<{ maxGuests: number; maxQueue: number; languages: string[] }>('/api/config'),

  createRoom: () => request<CreateRoomResponse>('/api/rooms', post({})),

  resolveRoom: (code: string) =>
    request<{ code: string }>(`/api/rooms/resolve?code=${encodeURIComponent(code)}`),

  getRoom: (code: string) => request<Snapshot>(`/api/rooms/${encodeURIComponent(code)}`),

  search: (q: string) =>
    request<{ results: SearchResult[]; stale: boolean }>(
      `/api/search?q=${encodeURIComponent(q)}`
    ),

  addSong: (code: string, payload: { videoId?: string; url?: string; nickname: string }) =>
    request<Snapshot>(`/api/rooms/${encodeURIComponent(code)}/songs`, post(payload)),

  toggleVote: (code: string, entryId: string) =>
    request<Snapshot>(
      `/api/rooms/${encodeURIComponent(code)}/songs/${encodeURIComponent(entryId)}/vote`,
      post({})
    ),

  startSkipVote: (code: string) =>
    request<Snapshot>(`/api/rooms/${encodeURIComponent(code)}/skip-vote`, post({})),

  castSkipVote: (code: string) =>
    request<Snapshot>(`/api/rooms/${encodeURIComponent(code)}/skip-vote/cast`, post({})),

  cancelSkipVote: (code: string) =>
    request<Snapshot>(`/api/rooms/${encodeURIComponent(code)}/skip-vote/cancel`, post({})),

  unlockHost: (code: string, pin: string) =>
    request<Snapshot>(`/api/rooms/${encodeURIComponent(code)}/host/unlock`, post({ pin })),

  playback: (code: string, action: 'play' | 'pause' | 'next' | 'cancelCountdown') =>
    request<Snapshot>(`/api/rooms/${encodeURIComponent(code)}/playback`, post({ action })),

  removeSong: (code: string, entryId: string) =>
    request<Snapshot>(
      `/api/rooms/${encodeURIComponent(code)}/songs/${encodeURIComponent(entryId)}`,
      { method: 'DELETE' }
    ),

  reorder: (code: string, entryIds: string[]) =>
    request<Snapshot>(`/api/rooms/${encodeURIComponent(code)}/order`, {
      method: 'PUT',
      body: JSON.stringify({ entryIds })
    }),

  clearQueue: (code: string) =>
    request<Snapshot>(`/api/rooms/${encodeURIComponent(code)}/queue/clear`, post({})),

  lockQueue: (code: string, locked: boolean) =>
    request<Snapshot>(`/api/rooms/${encodeURIComponent(code)}/queue/lock`, post({ locked })),

  rotatePIN: (code: string) =>
    request<{ pin: string }>(`/api/rooms/${encodeURIComponent(code)}/pin/rotate`, post({})),

  endRoom: (code: string) =>
    request<void>(`/api/rooms/${encodeURIComponent(code)}/end`, post({})),

  tvSession: () => request<TVSession>('/api/tv/session')
};
