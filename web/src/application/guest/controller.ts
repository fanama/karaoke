import { get } from 'svelte/store';
import { errorCodeOf } from '../../domain/errors';
import { tr } from '../i18n';
import { createRoomSession } from '../room_session';
import type { AppDependencies } from '../../ports/context';
import { createGuestActions } from './actions';
import { createGuestSearch } from './search';
import { createGuestState } from './state';

export function createGuestController(code: string, deps: AppDependencies) {
  const state = createGuestState();
  const session = createRoomSession({
    api: deps.api,
    connectSocket: deps.connectSocket,
    code: () => code,
    client: 'phone',
    translate: () => get(tr)
  });
  const ctx = { state, deps, session, code };
  const search = createGuestSearch(state, deps, session, code);
  const actions = createGuestActions(ctx);

  async function load(): Promise<void> {
    void search.shuffle();
    try {
      session.apply(await deps.api.getRoom(code));
      session.connect();
    } catch (err) {
      state.loadError.set(errorCodeOf(err, 'connection'));
    }
  }

  return {
    ...state,
    ...search,
    ...actions,
    snap: session.snap,
    status: session.status,
    ended: session.ended,
    toast: session.toast,
    load
  };
}
