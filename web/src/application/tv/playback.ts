import { get } from 'svelte/store';
import type { PlaybackAction } from '../../domain/models';
import type { AppDependencies } from '../../ports/context';
import type { RoomSession } from '../room_session';
import type { TvState } from './state';

export interface PlayerHandle {
  forcePlay(): void;
  setMuted(value: boolean): void;
}

export function createTvPlayback(state: TvState, deps: AppDependencies, session: RoomSession) {
  const playback = new TvPlayback(state, deps, session);
  return {
    setPlayer: (player: PlayerHandle | null) => playback.setPlayer(player),
    playback: (action: PlaybackAction) => playback.play(action),
    toggleMute: () => playback.toggleMute(),
    turnSoundOn: () => playback.turnSoundOn(),
    onSoundBlocked: (blocked: boolean) => playback.onSoundBlocked(blocked),
    onPlayerBlocked: (blocked: boolean) => playback.onPlayerBlocked(blocked)
  };
}

class TvPlayback {
  private player: PlayerHandle | null = null;

  constructor(
    private readonly state: TvState,
    private readonly deps: AppDependencies,
    private readonly session: RoomSession
  ) {}

  setPlayer(player: PlayerHandle | null): void {
    this.player = player;
    if (player) player.setMuted(get(this.state.muted));
  }

  async play(action: PlaybackAction): Promise<void> {
    if (get(this.state.busy)) return;
    this.state.busy.set(true);
    if (action !== 'pause') this.state.playerBlocked.set(false);
    try {
      this.session.apply(await this.deps.api.playback(get(this.state.code), action));
      if (action === 'play') this.player?.forcePlay();
    } catch (err) {
      this.session.showError(err);
    } finally {
      this.state.busy.set(false);
    }
  }

  toggleMute(): void {
    const next = !get(this.state.muted);
    this.state.muted.set(next);
    if (!next) this.state.soundBlocked.set(false);
    this.player?.setMuted(next);
  }

  turnSoundOn(): void {
    this.state.soundBlocked.set(false);
    this.state.muted.set(false);
    this.player?.setMuted(false);
  }

  onSoundBlocked(blocked: boolean): void {
    this.state.soundBlocked.set(blocked);
    this.state.muted.set(blocked);
  }

  onPlayerBlocked(blocked: boolean): void {
    this.state.playerBlocked.set(blocked);
  }
}
