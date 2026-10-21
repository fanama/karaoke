import type { Snapshot } from './models';

export function formatDuration(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds <= 0) return '—';
  const s = Math.floor(seconds % 60);
  const m = Math.floor((seconds / 60) % 60);
  const h = Math.floor(seconds / 3600);
  if (h > 0) return `${h}:${pad(m)}:${pad(s)}`;
  return `${m}:${pad(s)}`;
}

function pad(value: number): string {
  return String(value).padStart(2, '0');
}

export function secondsUntil(iso: string | null | undefined, now = Date.now()): number {
  if (!iso) return 0;
  return Math.max(0, Math.ceil((new Date(iso).getTime() - now) / 1000));
}

export function clampPercent(value: number): number {
  if (!Number.isFinite(value)) return 0;
  return Math.min(100, Math.max(0, value));
}

/** Elapsed playback time, extrapolated from the server snapshot while playing. */
export function computeElapsed(snap: Snapshot, now = Date.now()): number {
  const np = snap.nowPlaying;
  if (!np) return 0;
  const drift = np.status === 'playing' ? Math.max(0, (now - new Date(snap.serverTime).getTime()) / 1000) : 0;
  return np.elapsedSeconds + drift;
}

export function computeProgress(elapsed: number, duration: number): number {
  if (!duration) return 0;
  return clampPercent((elapsed / duration) * 100);
}

export function voteProgress(yes: number, threshold: number): number {
  return clampPercent((yes / Math.max(1, threshold)) * 100);
}
