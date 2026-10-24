CREATE TABLE IF NOT EXISTS songs (
  video_id         TEXT PRIMARY KEY,
  title            TEXT NOT NULL,
  channel          TEXT NOT NULL DEFAULT '',
  thumbnail_url    TEXT NOT NULL DEFAULT '',
  duration_seconds INTEGER NOT NULL DEFAULT 0,
  search_text      TEXT NOT NULL DEFAULT '',
  first_seen_at    TEXT NOT NULL,
  last_seen_at     TEXT NOT NULL
);

-- Search runs in the Render in-memory cache, not in D1. Deliberately no
-- secondary index here: it would add a written row to every upsert.
