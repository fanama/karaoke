import type { ApiFailure } from '../domain/errors';
import type { PlaybackAction } from '../domain/models';
import type { RoomApiPort } from '../ports/api_port';

export class ApiError extends Error implements ApiFailure {
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

type FetchLike = typeof fetch;
type RoomMethods = Omit<RoomApiPort, 'adminSession' | 'adminLogin' | 'adminLogout' | 'adminImport'>;
type AdminMethods = Pick<RoomApiPort, 'adminSession' | 'adminLogin' | 'adminLogout' | 'adminImport'>;

export function createApiAdapter(fetcher: FetchLike = fetch): RoomApiPort {
  return { ...roomMethods(fetcher), ...adminMethods(fetcher) };
}

function roomMethods(fetcher: FetchLike): RoomMethods {
  return {
    config: () => request(fetcher, '/api/config'),
    createRoom: () => request(fetcher, '/api/rooms', post({})),
    resolveRoom: (code) => request(fetcher, `/api/rooms/resolve?code=${enc(code)}`),
    getRoom: (code) => request(fetcher, `/api/rooms/${enc(code)}`),
    search: (query) => request(fetcher, `/api/search?q=${enc(query)}`),
    randomSongs: (limit = 6) => request(fetcher, `/api/songs/random?limit=${limit}`),
    libraryStats: (code) => request(fetcher, `/api/rooms/${enc(code)}/library`),
    importLibrary: (code, file) =>
      request(fetcher, `/api/rooms/${enc(code)}/library/import`, {
        method: 'POST',
        body: file,
        headers: { 'Content-Type': 'application/octet-stream' }
      }),
    addSong: (code, payload) => request(fetcher, `/api/rooms/${enc(code)}/songs`, post(payload)),
    toggleVote: (code, entryId) =>
      request(fetcher, `/api/rooms/${enc(code)}/songs/${enc(entryId)}/vote`, post({})),
    startSkipVote: (code) => request(fetcher, `/api/rooms/${enc(code)}/skip-vote`, post({})),
    castSkipVote: (code) => request(fetcher, `/api/rooms/${enc(code)}/skip-vote/cast`, post({})),
    cancelSkipVote: (code) => request(fetcher, `/api/rooms/${enc(code)}/skip-vote/cancel`, post({})),
    unlockHost: (code, pin) => request(fetcher, `/api/rooms/${enc(code)}/host/unlock`, post({ pin })),
    playback: (code, action: PlaybackAction) =>
      request(fetcher, `/api/rooms/${enc(code)}/playback`, post({ action })),
    removeSong: (code, entryId) =>
      request(fetcher, `/api/rooms/${enc(code)}/songs/${enc(entryId)}`, { method: 'DELETE' }),
    reorder: (code, entryIds) =>
      request(fetcher, `/api/rooms/${enc(code)}/order`, {
        method: 'PUT',
        body: JSON.stringify({ entryIds })
      }),
    clearQueue: (code) => request(fetcher, `/api/rooms/${enc(code)}/queue/clear`, post({})),
    lockQueue: (code, locked) => request(fetcher, `/api/rooms/${enc(code)}/queue/lock`, post({ locked })),
    rotatePIN: (code) => request(fetcher, `/api/rooms/${enc(code)}/pin/rotate`, post({})),
    endRoom: (code) => request(fetcher, `/api/rooms/${enc(code)}/end`, post({})),
    tvSession: () => request(fetcher, '/api/tv/session')
  };
}

function adminMethods(fetcher: FetchLike): AdminMethods {
  return {
    adminSession: () => request(fetcher, '/api/admin/session'),
    adminLogin: (token: string) => request(fetcher, '/api/admin/login', post({ token })),
    adminLogout: () => request(fetcher, '/api/admin/logout', post({})),
    adminImport: (file: File, format: 'sqlite' | 'sql') =>
      request(fetcher, `/api/admin/library/import?format=${format}`, {
        method: 'POST',
        body: file,
        headers: { 'Content-Type': 'application/octet-stream' }
      })
  };
}

const enc = encodeURIComponent;

const post = (body?: unknown): RequestInit => ({
  method: 'POST',
  body: body === undefined ? undefined : JSON.stringify(body)
});

async function request<T>(fetcher: FetchLike, path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json');
  const res = await fetcher(path, { ...init, headers, credentials: 'same-origin' });
  if (!res.ok) throw await failure(res);
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

async function failure(res: Response): Promise<ApiError> {
  let body: any = null;
  try {
    body = await res.json();
  } catch {
    // non-JSON error body
  }
  return new ApiError(
    body?.error?.code ?? 'internal_error',
    body?.error?.message ?? res.statusText,
    res.status,
    Number(body?.error?.retry_after ?? res.headers.get('Retry-After') ?? 0)
  );
}
