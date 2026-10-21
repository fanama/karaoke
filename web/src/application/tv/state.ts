import { writable } from 'svelte/store';

export function createTvState() {
  return {
    code: writable(''),
    pin: writable(''),
    badLink: writable(false),
    muted: writable(false),
    playerBlocked: writable(false),
    soundBlocked: writable(false),
    queueOpen: writable(false),
    busy: writable(false),
    confirmClear: writable(false),
    confirmEnd: writable(false),
    dragId: writable('')
  };
}

export type TvState = ReturnType<typeof createTvState>;
