import { writable } from 'svelte/store';
import type { QueueSelection, SearchResult } from '../../domain/models';

export function createGuestState() {
  return {
    loadError: writable(''),
    tab: writable<'search' | 'queue'>('search'),
    q: writable(''),
    results: writable<SearchResult[]>([]),
    searching: writable(false),
    searched: writable(false),
    stale: writable(false),
    searchFailed: writable(false),
    selected: writable<QueueSelection | null>(null),
    nickname: writable(''),
    nicknameError: writable(''),
    adding: writable(false),
    addError: writable(''),
    pasteUrl: writable(''),
    pasteError: writable(''),
    pinSheet: writable(false),
    pin: writable(''),
    pinError: writable(''),
    unlocking: writable(false),
    manageOpen: writable(false),
    confirmClear: writable(false),
    confirmEnd: writable(false),
    busy: writable(false)
  };
}

export type GuestState = ReturnType<typeof createGuestState>;
