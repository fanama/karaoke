export interface D1Result<T = unknown> {
  success: boolean;
  results?: T[];
  meta?: {
    rows_read?: number;
    rows_written?: number;
    changes?: number;
  };
}

export interface D1PreparedStatement {
  bind(...values: unknown[]): D1PreparedStatement;
  all<T = unknown>(): Promise<D1Result<T>>;
  run<T = unknown>(): Promise<D1Result<T>>;
}

export interface D1Database {
  prepare(sql: string): D1PreparedStatement;
  batch(statements: D1PreparedStatement[]): Promise<D1Result[]>;
}

export interface Env {
  DB: D1Database;
  LIBRARY_PROXY_TOKEN: string;
}

export interface SongInput {
  videoId?: unknown;
  title?: unknown;
  channel?: unknown;
  thumbnailUrl?: unknown;
  durationSeconds?: unknown;
  searchText?: unknown;
  firstSeenAt?: unknown;
  lastSeenAt?: unknown;
}

export interface SongRow {
  video_id: string;
  title: string;
  channel: string;
  thumbnail_url: string;
  duration_seconds: number;
  search_text: string;
  first_seen_at: string;
  last_seen_at: string;
}

export interface CleanSong {
  videoId: string;
  title: string;
  channel: string;
  thumbnailUrl: string;
  durationSeconds: number;
  searchText: string;
  firstSeenAt: string;
  lastSeenAt: string;
}
