import { HttpError, json } from './http';
import type { D1Database, SongRow } from './types';

const LIST_SQL = `SELECT video_id, title, channel, thumbnail_url, duration_seconds,
  search_text, first_seen_at, last_seen_at FROM songs
  ORDER BY video_id LIMIT ? OFFSET ?`;

export async function listSongs(request: Request, db: D1Database): Promise<Response> {
  const url = new URL(request.url);
  const limit = integerParam(url.searchParams.get('limit'), 500, 1, 500);
  const offset = integerParam(url.searchParams.get('offset'), 0, 0, 1_000_000);
  const result = await db.prepare(LIST_SQL).bind(limit, offset).all<SongRow>();
  const rows = result.results ?? [];
  const body: Record<string, unknown> = { songs: rows.map(rowToSong) };
  if (rows.length === limit) body.nextOffset = offset + limit;
  return json(body);
}

function integerParam(raw: string | null, fallback: number, min: number, max: number): number {
  if (raw === null || raw === '') return fallback;
  const value = Number(raw);
  if (!Number.isInteger(value) || value < min || value > max) {
    throw new HttpError(400, 'invalid_request', 'Invalid pagination parameter.');
  }
  return value;
}

function rowToSong(row: SongRow): Record<string, unknown> {
  return {
    videoId: row.video_id,
    title: row.title,
    channel: row.channel,
    thumbnailUrl: row.thumbnail_url,
    durationSeconds: row.duration_seconds,
    searchText: row.search_text,
    firstSeenAt: row.first_seen_at,
    lastSeenAt: row.last_seen_at
  };
}
