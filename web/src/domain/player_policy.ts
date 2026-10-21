/** How often the player checks whether playback actually started. */
export const BLOCKED_CHECK_MS = 750;
/** How long the player keeps nudging a stalled iframe before asking for a tap. */
export const BLOCKED_GIVE_UP_MS = 18_000;
/** A drift larger than this is corrected with a seek. */
export const DRIFT_TOLERANCE_SECONDS = 2;
/** Muted autoplay fallbacks before the player gives up and asks for a gesture. */
export const MAX_SOUND_RECOVERIES = 3;
/** Playback telemetry is flushed at this interval. */
export const PROGRESS_INTERVAL_MS = 5000;
/** Reconnection backoff, capped. */
export const RECONNECT_BASE_MS = 500;
export const RECONNECT_MAX_MS = 10_000;

export type AutoplayDecision = 'play' | 'mute' | 'blocked' | 'wait';

/**
 * Decides what to do when playback does not visibly progress.
 * Pure so the browser-specific player can stay thin.
 */
export function autoplayDecision(input: {
  attempt: number;
  position: number;
  mutedByAutoplay: boolean;
  soundUnlocked: boolean;
  recoveries: number;
}): AutoplayDecision {
  // Only the media clock proves playback: a blocked iframe can report
  // "playing" (state 1) while its position stays frozen at zero.
  if (input.position > 0.5) return 'wait';
  const waited = (input.attempt + 1) * BLOCKED_CHECK_MS;
  const canFallback = !input.mutedByAutoplay && !input.soundUnlocked && input.recoveries < MAX_SOUND_RECOVERIES;
  if (canFallback && waited >= 2 * BLOCKED_CHECK_MS) return 'mute';
  if (waited < BLOCKED_GIVE_UP_MS) return 'play';
  return 'blocked';
}
