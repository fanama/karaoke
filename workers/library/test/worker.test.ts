import { describe, expect, it } from 'vitest';
import worker from '../src/index';
import type { D1Database, D1PreparedStatement, D1Result } from '../src/types';

const TOKEN = 'test-token';

class FakeStatement implements D1PreparedStatement {
  values: unknown[] = [];
  constructor(
    private readonly db: FakeDB,
    readonly sql: string
  ) {}

  bind(...values: unknown[]): D1PreparedStatement {
    this.values = values;
    return this;
  }

  all<T = unknown>(): Promise<D1Result<T>> {
    return Promise.resolve({ success: true, results: this.db.rows as T[] });
  }

  run<T = unknown>(): Promise<D1Result<T>> {
    return Promise.resolve({ success: true });
  }
}

class FakeDB implements D1Database {
  rows: unknown[] = [];
  statements: FakeStatement[] = [];
  batches = 0;

  prepare(sql: string): D1PreparedStatement {
    return new FakeStatement(this, sql);
  }

  batch(statements: D1PreparedStatement[]): Promise<D1Result[]> {
    this.batches++;
    this.statements.push(...(statements as FakeStatement[]));
    return Promise.resolve(statements.map(() => ({ success: true })));
  }
}

function request(path: string, init: RequestInit = {}): Request {
  return new Request(`https://worker.test${path}`, init);
}

function post(path: string, body: unknown): Request {
  return request(path, {
    method: 'POST',
    headers: { Authorization: `Bearer ${TOKEN}`, 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
}

function song(id: string, title = 'Karaoke') {
  return { videoId: id, title, channel: 'Channel', durationSeconds: 200 };
}

function env(db = new FakeDB()) {
  return { DB: db, LIBRARY_PROXY_TOKEN: TOKEN };
}

describe('worker', () => {
  it('rejects a missing bearer token', async () => {
    const res = await worker.fetch(request('/v1/health'), env());
    expect(res.status).toBe(401);
  });

  it('reports health', async () => {
    const res = await worker.fetch(request('/v1/health', { headers: { Authorization: `Bearer ${TOKEN}` } }), env());
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({ ok: true });
  });

  it('upserts a validated batch in one D1 batch', async () => {
    const db = new FakeDB();
    const res = await worker.fetch(post('/v1/songs/upsert', { songs: [song('aaaaaaaaaaa'), song('bbbbbbbbbbb')] }), env(db));
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({ written: 2 });
    expect(db.batches).toBe(1);
    expect(db.statements[0].values).toHaveLength(16);
  });

  it('rejects an invalid video id', async () => {
    const res = await worker.fetch(post('/v1/songs/upsert', { songs: [song('short')] }), env());
    expect(res.status).toBe(400);
  });

  it('returns a stable page and next offset', async () => {
    const db = new FakeDB();
    db.rows = [{ video_id: 'aaaaaaaaaaa', title: 'A', channel: '', thumbnail_url: '', duration_seconds: 1, search_text: 'a', first_seen_at: '', last_seen_at: '' }];
    const res = await worker.fetch(request('/v1/songs?limit=1&offset=0', { headers: { Authorization: `Bearer ${TOKEN}` } }), env(db));
    expect(res.status).toBe(200);
    const body = (await res.json()) as { songs: unknown[]; nextOffset?: number };
    expect(body.songs).toHaveLength(1);
    expect(body.nextOffset).toBe(1);
  });

  it('imports more than one chunk in a single batch call', async () => {
    const db = new FakeDB();
    const songs = Array.from({ length: 11 }, (_, i) => song(`aaaaaaaaaa${String.fromCharCode(97 + i)}`));
    const res = await worker.fetch(post('/v1/songs/import', { songs }), env(db));
    expect(res.status).toBe(200);
    expect(db.batches).toBe(1);
    expect(db.statements).toHaveLength(2);
  });
});
