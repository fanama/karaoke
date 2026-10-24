import { failure, HttpError } from './http';

export function errorResponse(err: unknown): Response {
  if (err instanceof HttpError) return failure(err.status, err.code, err.message);
  const message = err instanceof Error ? err.message : '';
  if (message.includes('daily row read limit') || message.includes('daily row write limit')) {
    return failure(503, 'd1_limit_reached', 'Database daily limit reached.');
  }
  console.error('worker request failed');
  return failure(503, 'd1_unavailable', 'Database unavailable.');
}
