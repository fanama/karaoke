import { get } from 'svelte/store';
import { tr } from '../i18n';
import { createRoomSession } from '../room_session';
import type { AppDependencies } from '../../ports/context';
import { createTvHostActions } from './host';
import { createTvPlayback } from './playback';
import { createTvState } from './state';

export type { PlayerHandle } from './playback';

export function createTvController(token: string, deps: AppDependencies) {
  const state = createTvState();
  const session = createRoomSession({
    api: deps.api,
    connectSocket: deps.connectSocket,
    code: () => get(state.code),
    client: 'tv',
    translate: () => get(tr)
  });
  const host = createTvHostActions({ state, deps, session });
  const playback = createTvPlayback(state, deps, session);
  const room = new TvRoom(token, deps, state, session);

  return {
    ...state,
    ...host,
    ...playback,
    snap: session.snap,
    status: session.status,
    ended: session.ended,
    toast: session.toast,
    boot: () => room.boot(),
    telemetry: (type: string, payload: Record<string, unknown>) => session.send({ type, ...payload }),
    copyLink: () => room.copyLink(),
    newRoom: () => room.newRoom(),
    fullscreen: () => deps.fullscreen.toggle(),
    dispose: () => {
      session.disconnect();
      host.dispose();
    }
  };
}

class TvRoom {
  constructor(
    private readonly token: string,
    private readonly deps: AppDependencies,
    private readonly state: ReturnType<typeof createTvState>,
    private readonly session: ReturnType<typeof createRoomSession>
  ) {}

  async boot(): Promise<void> {
    if (!this.token) {
      this.state.badLink.set(true);
      return;
    }
    try {
      const tv = await this.deps.api.tvSession();
      this.state.code.set(tv.code);
      this.state.pin.set(tv.pin);
      this.session.apply(await this.deps.api.getRoom(tv.code));
      this.session.connect();
      document.title = `${tv.code.toUpperCase()} · Fana Karaoke`;
    } catch {
      this.state.badLink.set(true);
    }
  }

  async copyLink(): Promise<void> {
    try {
      await this.deps.clipboard.writeText(location.href);
      this.session.showToast(get(tr)('tv.tvLinkCopied'));
    } catch {
      this.session.showToast(location.href);
    }
  }

  async newRoom(): Promise<void> {
    try {
      const room = await this.deps.api.createRoom();
      this.deps.navigation.assign(room.tvUrl);
    } catch (err) {
      this.session.showError(err);
    }
  }
}

export type TvController = ReturnType<typeof createTvController>;
