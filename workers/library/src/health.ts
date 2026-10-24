import { json } from './http';
import type { D1Database } from './types';

export async function health(db: D1Database): Promise<Response> {
  await db.prepare('SELECT 1 AS ok').all();
  return json({ ok: true });
}
