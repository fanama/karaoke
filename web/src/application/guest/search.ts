import { get } from 'svelte/store';
import { errorCodeOf } from '../../domain/errors';
import type { SearchResult } from '../../domain/models';
import { isValidNickname, looksLikeYouTubeUrl } from '../../domain/room_code';
import { errorText, tr } from '../i18n';
import type { AppDependencies } from '../../ports/context';
import type { RoomSession } from '../room_session';
import type { GuestState } from './state';

export function createGuestSearch(state: GuestState, deps: AppDependencies, session: RoomSession, code: string) {
  const actions = new GuestSearch(state, deps, session, code);
  return {
    queue: () => actions.queue(),
    run: () => actions.run(),
    clear: () => actions.clear(),
    choose: (result: SearchResult) => actions.choose(result),
    preparePaste: () => actions.preparePaste(),
    submit: () => actions.submit()
  };
}

class GuestSearch {
  private timer: ReturnType<typeof setTimeout> | null = null;
  private seq = 0;

  constructor(
    private readonly state: GuestState,
    private readonly deps: AppDependencies,
    private readonly session: RoomSession,
    private readonly code: string
  ) {}

  queue(): void {
    if (this.timer) clearTimeout(this.timer);
    this.timer = setTimeout(() => this.run(), 350);
  }

  async run(): Promise<void> {
    const value = get(this.state.q).trim();
    if (value.length < 2) return this.reset();
    const current = ++this.seq;
    this.state.searching.set(true);
    this.state.searchFailed.set(false);
    try {
      const res = await this.deps.api.search(value);
      if (current !== this.seq) return;
      this.state.results.set(res.results);
      this.state.stale.set(res.stale);
      this.state.searched.set(true);
    } catch {
      if (current !== this.seq) return;
      this.state.results.set([]);
      this.state.searched.set(true);
      this.state.searchFailed.set(true);
    } finally {
      if (current === this.seq) this.state.searching.set(false);
    }
  }

  clear(): void {
    this.state.q.set('');
    this.state.stale.set(false);
    this.reset();
    if (this.timer) clearTimeout(this.timer);
  }

  choose(result: SearchResult): void {
    this.state.selected.set({
      videoId: result.videoId,
      title: result.title,
      channel: result.channel,
      thumbnailUrl: result.thumbnailUrl
    });
    this.startAdd();
  }

  preparePaste(): void {
    const raw = get(this.state.pasteUrl).trim();
    this.state.pasteError.set('');
    if (!raw || !looksLikeYouTubeUrl(raw)) {
      this.state.pasteError.set(get(tr)('error.invalid_url'));
      return;
    }
    this.state.selected.set({ url: raw, title: raw });
    this.startAdd();
  }

  async submit(): Promise<void> {
    const chosen = get(this.state.selected);
    if (!chosen || get(this.state.adding)) return;
    const name = get(this.state.nickname).trim();
    if (!isValidNickname(name)) {
      this.state.nicknameError.set(get(tr)('room.nicknameError'));
      return;
    }
    this.state.adding.set(true);
    this.state.addError.set('');
    try {
      const snap = await this.deps.api.addSong(this.code, { nickname: name, ...chosen });
      this.session.apply(snap);
      this.state.selected.set(null);
      this.state.nickname.set('');
      this.state.pasteUrl.set('');
      this.state.tab.set('queue');
      this.session.showToast(get(tr)('room.added'));
    } catch (err) {
      this.state.addError.set(errorText(get(tr), errorCodeOf(err, 'video_unavailable')));
    } finally {
      this.state.adding.set(false);
    }
  }

  private reset(): void {
    this.state.results.set([]);
    this.state.searched.set(false);
    this.state.searchFailed.set(false);
    this.state.searching.set(false);
  }

  private startAdd(): void {
    this.state.nickname.set('');
    this.state.nicknameError.set('');
    this.state.addError.set('');
  }
}
