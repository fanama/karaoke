import { get } from 'svelte/store';
import type { Snapshot } from '../../domain/models';
import { tr } from '../i18n';
import type { AppDependencies } from '../../ports/context';
import type { RoomSession } from '../room_session';
import type { TvState } from './state';

export interface TvHostContext {
  state: TvState;
  deps: AppDependencies;
  session: RoomSession;
}

export function createTvHostActions(ctx: TvHostContext) {
  const actions = new TvHostActions(ctx);
  return {
    toggleLock: () => actions.toggleLock(),
    removeSong: (id: string) => actions.removeSong(id),
    clearQueue: () => actions.clearQueue(),
    rotatePin: () => actions.rotatePin(),
    endRoom: () => actions.endRoom(),
    armEnd: () => actions.armEnd(),
    dragStart: (event: DragEvent, id: string) => actions.dragStart(event, id),
    drop: (event: DragEvent, id: string) => actions.drop(event, id),
    cancelVote: () => actions.cancelVote(),
    dispose: () => actions.dispose()
  };
}

class TvHostActions {
  private endTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(private readonly ctx: TvHostContext) {}

  private get code(): string {
    return get(this.ctx.state.code);
  }

  async toggleLock(): Promise<void> {
    const snap = get(this.ctx.session.snap);
    if (!snap) return;
    await this.mutate(() => this.ctx.deps.api.lockQueue(this.code, !snap.room.queueLocked));
  }

  async removeSong(entryId: string): Promise<void> {
    await this.mutate(() => this.ctx.deps.api.removeSong(this.code, entryId));
  }

  async clearQueue(): Promise<void> {
    await this.mutate(() => this.ctx.deps.api.clearQueue(this.code));
    this.ctx.state.confirmClear.set(false);
  }

  async rotatePin(): Promise<void> {
    try {
      const res = await this.ctx.deps.api.rotatePIN(this.code);
      this.ctx.state.pin.set(res.pin);
      this.ctx.session.showToast(get(tr)('room.pinChanged'));
    } catch (err) {
      this.ctx.session.showError(err);
    }
  }

  async endRoom(): Promise<void> {
    try {
      await this.ctx.deps.api.endRoom(this.code);
      this.ctx.session.disconnect();
      this.ctx.session.ended.set(true);
    } catch (err) {
      this.ctx.session.showError(err);
    }
  }

  armEnd(): void {
    if (get(this.ctx.state.confirmEnd)) {
      this.endRoom();
      return;
    }
    this.ctx.state.confirmEnd.set(true);
    if (this.endTimer) clearTimeout(this.endTimer);
    this.endTimer = setTimeout(() => this.ctx.state.confirmEnd.set(false), 3500);
  }

  dragStart(event: DragEvent, id: string): void {
    this.ctx.state.dragId.set(id);
    event.dataTransfer?.setData('text/plain', id);
    if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move';
  }

  drop(event: DragEvent, targetId: string): void {
    event.preventDefault();
    const dragId = get(this.ctx.state.dragId);
    if (!dragId || dragId === targetId) return;
    const ids = (get(this.ctx.session.snap)?.queue ?? []).map((item) => item.id);
    const from = ids.indexOf(dragId);
    const to = ids.indexOf(targetId);
    if (from < 0 || to < 0) return;
    ids.splice(from, 1);
    ids.splice(to, 0, dragId);
    this.ctx.state.dragId.set('');
    this.reorder(ids);
  }

  async cancelVote(): Promise<void> {
    try {
      this.ctx.session.apply(await this.ctx.deps.api.cancelSkipVote(this.code));
    } catch (err) {
      this.ctx.session.showError(err);
    }
  }

  dispose(): void {
    if (this.endTimer) clearTimeout(this.endTimer);
  }

  private async reorder(ids: string[]): Promise<void> {
    await this.mutate(() => this.ctx.deps.api.reorder(this.code, ids));
  }

  private async mutate(call: () => Promise<Snapshot>): Promise<void> {
    try {
      this.ctx.session.apply(await call());
    } catch (err) {
      this.ctx.session.showError(err);
    }
  }
}
