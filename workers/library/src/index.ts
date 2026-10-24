import { authorized } from './auth';
import { errorResponse } from './errors';
import { failure } from './http';
import { health } from './health';
import { listSongs } from './songs';
import { importSongs, upsertSongs } from './upsert';
import type { Env } from './types';

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    try {
      return await route(request, env);
    } catch (err) {
      return errorResponse(err);
    }
  }
};

async function route(request: Request, env: Env): Promise<Response> {
  if (!authorized(request, env.LIBRARY_PROXY_TOKEN)) {
    return failure(401, 'unauthorized', 'Unauthorized.');
  }
  const { pathname } = new URL(request.url);
  if (request.method === 'GET' && pathname === '/v1/health') return health(env.DB);
  if (request.method === 'GET' && pathname === '/v1/songs') return listSongs(request, env.DB);
  if (request.method === 'POST' && pathname === '/v1/songs/upsert') return upsertSongs(request, env);
  if (request.method === 'POST' && pathname === '/v1/songs/import') return importSongs(request, env);
  return failure(404, 'not_found', 'Not found.');
}
