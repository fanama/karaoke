import { get } from 'svelte/store';
import { errorCodeOf } from '../../domain/errors';
import type { PlaybackAction, Snapshot } from '../../domain/models';
import { errorText, tr } from '../i18n';
import type { AppDependencies } from '../../ports/context';
import type { RoomSession } from '../room_session';
import type { GuestState } from './state';

export interface GuestActionContext {
  state: GuestState;
  deps: AppDependencies;
  session: RoomSession;
  code: string;
}

export function createGuestActions(ctx: GuestActionContext) {
  const actions = new GuestActions(ctx);
  return {
    toggleVote: (entryId: string) => actions.toggleVote(entryId),
    startSkip: () => actions.startSkip(),
    castSkip: () => actions.castSkip(),
    cancelSkip: () => actions.cancelSkip(),
    removeSong: (entryId: string) => actions.removeSong(entryId),
    playback: (action: PlaybackAction) => actions.playback(action),
    moveSong: (index: number, direction: -1 | 1) => actions.moveSong(index, direction),
    unlock: () => actions.unlock(),
    toggleLock: () => actions.toggleLock(),
    clearQueue: () => actions.clearQueue(),
    loadLibrary: () => actions.loadLibrary(),
    importLibrary: (file: File) => actions.importLibrary(file),
    endRoom: () => actions.endRoom(),
    newRoom: () => actions.newRoom()
  };
}

class GuestActions {
  constructor(private readonly ctx: GuestActionContext) {}

  async toggleVote(entryId: string): Promise<void> {
    await this.mutate(() => this.ctx.deps.api.toggleVote(this.ctx.code, entryId));
  }

  async startSkip(): Promise<void> {
    await this.mutate(() => this.ctx.deps.api.startSkipVote(this.ctx.code));
  }

  async castSkip(): Promise<void> {
    await this.mutate(() => this.ctx.deps.api.castSkipVote(this.ctx.code));
  }

  async cancelSkip(): Promise<void> {
    await this.mutate(() => this.ctx.deps.api.cancelSkipVote(this.ctx.code));
  }

  async removeSong(entryId: string): Promise<void> {
    await this.mutate(() => this.ctx.deps.api.removeSong(this.ctx.code, entryId));
  }

  async playback(action: PlaybackAction): Promise<void> {
    if (get(this.ctx.state.busy)) return;
    this.ctx.state.busy.set(true);
    try {
      await this.mutate(() => this.ctx.deps.api.playback(this.ctx.code, action));
    } finally {
      this.ctx.state.busy.set(false);
    }
  }

  async moveSong(index: number, direction: -1 | 1): Promise<void> {
    const queue = get(this.ctx.session.snap)?.queue ?? [];
    const target = index + direction;
    if (target < 0 || target >= queue.length) return;
    const ids = queue.map((item) => item.id);
    [ids[index], ids[target]] = [ids[target], ids[index]];
    await this.mutate(() => this.ctx.deps.api.reorder(this.ctx.code, ids));
  }

  async unlock(): Promise<void> {
    if (get(this.ctx.state.unlocking)) return;
    this.ctx.state.unlocking.set(true);
    this.ctx.state.pinError.set('');
    try {
      const snap = await this.ctx.deps.api.unlockHost(this.ctx.code, get(this.ctx.state.pin).trim());
      this.ctx.session.apply(snap);
      this.ctx.state.pinSheet.set(false);
      this.ctx.state.pin.set('');
      this.ctx.session.connect();
      this.ctx.session.showToast(get(tr)('notice.hostUnlocked'));
    } catch (err) {
      this.ctx.state.pinError.set(errorText(get(tr), errorCodeOf(err, 'invalid_pin')));
    } finally {
      this.ctx.state.unlocking.set(false);
    }
  }

  async toggleLock(): Promise<void> {
    const snap = get(this.ctx.session.snap);
    if (!snap) return;
    await this.mutate(() => this.ctx.deps.api.lockQueue(this.ctx.code, !snap.room.queueLocked));
  }

  async clearQueue(): Promise<void> {
    try {
      this.ctx.session.apply(await this.ctx.deps.api.clearQueue(this.ctx.code));
      this.ctx.state.confirmClear.set(false);
      this.ctx.state.manageOpen.set(false);
    } catch (err) {
      this.ctx.session.showError(err);
    }
  }

  async loadLibrary(): Promise<void> {
    try {
      const res = await this.ctx.deps.api.libraryStats(this.ctx.code);
      this.ctx.state.libraryCount.set(res.songs);
      this.ctx.state.libraryError.set('');
    } catch (err) {
      this.ctx.state.libraryCount.set(null);
      this.ctx.state.libraryError.set(errorText(get(tr), errorCodeOf(err, 'library_disabled')));
    }
  }

  async importLibrary(file: File): Promise<void> {
    if (get(this.ctx.state.libraryBusy)) return;
    this.ctx.state.libraryBusy.set(true);
    this.ctx.state.libraryError.set('');
    try {
      const res = await this.ctx.deps.api.importLibrary(this.ctx.code, file);
      this.ctx.state.libraryCount.set(res.songs);
      this.ctx.session.showToast(
        get(tr)('room.libraryImported', { count: res.imported, total: res.songs })
      );
    } catch (err) {
      this.ctx.state.libraryError.set(errorText(get(tr), errorCodeOf(err, 'library_invalid')));
    } finally {
      this.ctx.state.libraryBusy.set(false);
    }
  }

  async endRoom(): Promise<void> {
    try {
      await this.ctx.deps.api.endRoom(this.ctx.code);
      this.ctx.session.disconnect();
      this.ctx.session.ended.set(true);
      this.ctx.state.manageOpen.set(false);
    } catch (err) {
      this.ctx.session.showError(err);
    }
  }

  async newRoom(): Promise<void> {
    try {
      const room = await this.ctx.deps.api.createRoom();
      this.ctx.deps.navigation.assign(room.tvUrl);
    } catch (err) {
      this.ctx.session.showError(err, 'create');
    }
  }

  private async mutate(call: () => Promise<Snapshot>): Promise<void> {
    try {
      this.ctx.session.apply(await call());
    } catch (err) {
      this.ctx.session.showError(err);
    }
  }
}
