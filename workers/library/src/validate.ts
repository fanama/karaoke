import { HttpError } from './http';
import type { CleanSong, SongInput } from './types';

const VIDEO_ID = /^[A-Za-z0-9_-]{11}$/;

export function parseSongList(body: Record<string, unknown>, max: number): CleanSong[] {
  const raw = body.songs;
  if (!Array.isArray(raw) || raw.length < 1) {
    throw new HttpError(400, 'invalid_request', 'songs must be a non-empty array.');
  }
  if (raw.length > max) {
    throw new HttpError(413, 'too_large', `At most ${max} songs are allowed.`);
  }
  const songs = raw.map((item) => cleanSong(item));
  return dedupeSongs(songs);
}

export function cleanSong(value: unknown): CleanSong {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new HttpError(400, 'invalid_song', 'Invalid song payload.');
  }
  const song = value as SongInput;
  const videoId = requiredText(song.videoId, 'videoId', 11);
  if (!VIDEO_ID.test(videoId)) throw invalidSong();
  const title = requiredText(song.title, 'title', 300);
  const channel = optionalText(song.channel, 'channel', 200);
  const thumbnailUrl = optionalText(song.thumbnailUrl, 'thumbnailUrl', 500);
  const durationSeconds = integer(song.durationSeconds, 0, 36000);	const searchText = optionalText(song.searchText, 'searchText', 1000) || `${title} ${channel}`.toLowerCase();
	const now = new Date().toISOString();
	const firstSeenAt = timestamp(song.firstSeenAt) || now;
	const lastSeenAt = timestamp(song.lastSeenAt) || firstSeenAt;
  return { videoId, title, channel, thumbnailUrl, durationSeconds, searchText, firstSeenAt, lastSeenAt };
}

export function dedupeSongs(songs: CleanSong[]): CleanSong[] {
  const byId = new Map<string, CleanSong>();
  for (const song of songs) byId.set(song.videoId, song);
  return [...byId.values()];
}

export function chunkSongs(songs: CleanSong[], size = 10): CleanSong[][] {
  const chunks: CleanSong[][] = [];
  for (let i = 0; i < songs.length; i += size) chunks.push(songs.slice(i, i + size));
  return chunks;
}

function requiredText(value: unknown, field: string, max: number): string {
  const text = optionalText(value, field, max);
  if (!text) throw new HttpError(400, 'invalid_song', `${field} is required.`);
  return text;
}

function optionalText(value: unknown, field: string, max: number): string {
  if (value === undefined || value === null) return '';
  if (typeof value !== 'string') throw new HttpError(400, 'invalid_song', `${field} must be a string.`);
  const text = value.trim();
  if (text.length > max) throw new HttpError(400, 'invalid_song', `${field} is too long.`);
  return text;
}

function integer(value: unknown, min: number, max: number): number {
  const number = typeof value === 'number' ? value : Number(value ?? 0);
  if (!Number.isInteger(number) || number < min || number > max) {
    throw new HttpError(400, 'invalid_song', 'durationSeconds is invalid.');
  }
  return number;
}

function timestamp(value: unknown): string {
  if (value === undefined || value === null || value === '') return '';
  if (typeof value !== 'string' || value.length > 40) throw invalidSong();
  const parsed = Date.parse(value);
  if (!Number.isFinite(parsed)) throw invalidSong();
  return new Date(parsed).toISOString();
}

function invalidSong(): HttpError {
  return new HttpError(400, 'invalid_song', 'Invalid song payload.');
}
