/** Shape of an API failure, without importing the transport adapter. */
export interface ApiFailure {
  code: string;
  status: number;
  retryAfter: number;
  message?: string;
}

/** Reads the stable error code from anything the API layer throws. */
export function errorCodeOf(err: unknown, fallback = 'generic'): string {
  if (isApiFailure(err) && err.code) return err.code;
  return fallback;
}

export function isApiFailure(err: unknown): err is ApiFailure {
  return typeof err === 'object' && err !== null && typeof (err as ApiFailure).code === 'string';
}

export function isRoomGone(err: unknown): boolean {
  return errorCodeOf(err) === 'room_not_found';
}
