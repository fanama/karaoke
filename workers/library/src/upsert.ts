import { json, readObject } from './http';
import type { CleanSong, D1Database, D1PreparedStatement, Env } from './types';
import { chunkSongs, parseSongList } from './validate';

const UPSERT_PREFIX = `INSERT INTO songs (
  video_id, title, channel, thumbnail_url, duration_seconds,
  search_text, first_seen_at, last_seen_at
) VALUES `;

const UPSERT_SUFFIX = ` ON CONFLICT (video_id) DO UPDATE SET
  title = excluded.title,
  channel = excluded.channel,
  thumbnail_url = excluded.thumbnail_url,
  duration_seconds = excluded.duration_seconds,
  search_text = excluded.search_text,
  first_seen_at = songs.first_seen_at,
  last_seen_at = excluded.last_seen_at`;

export async function upsertSongs(request: Request, env: Env): Promise<Response> {
  return writeSongs(request, env, 100);
}

export async function importSongs(request: Request, env: Env): Promise<Response> {
  return writeSongs(request, env, 2000);
}

async function writeSongs(request: Request, env: Env, max: number): Promise<Response> {
  const body = await readObject(request);
  const songs = parseSongList(body, max);
  const statements = chunkSongs(songs).map((chunk) => statementFor(env.DB, chunk));
  await env.DB.batch(statements);
  return json({ written: songs.length });
}

function statementFor(db: D1Database, songs: CleanSong[]): D1PreparedStatement {
  const placeholders = songs.map(() => '(?, ?, ?, ?, ?, ?, ?, ?)').join(', ');
  const statement = db.prepare(UPSERT_PREFIX + placeholders + UPSERT_SUFFIX);
  return statement.bind(...songs.flatMap(songValues));
}

function songValues(song: CleanSong): unknown[] {
  return [
    song.videoId,
    song.title,
    song.channel,
    song.thumbnailUrl,
    song.durationSeconds,
    song.searchText,
    song.firstSeenAt,
    song.lastSeenAt
  ];
}
