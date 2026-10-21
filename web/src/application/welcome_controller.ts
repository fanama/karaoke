import { get, writable } from 'svelte/store';
import { errorCodeOf } from '../domain/errors';
import { normalizeCode } from '../domain/room_code';
import { errorText, tr } from './i18n';
import type { AppDependencies } from '../ports/context';

export function createWelcomeController(deps: AppDependencies) {
  const creating = writable(false);
  const createError = writable('');
  const code = writable('');
  const joinError = writable('');
  const joining = writable(false);

  async function createRoom(): Promise<void> {
    if (get(creating)) return;
    creating.set(true);
    createError.set('');
    try {
      const room = await deps.api.createRoom();
      deps.navigation.assign(room.tvUrl);
    } catch (err) {
      createError.set(errorText(get(tr), errorCodeOf(err, 'create')));
      creating.set(false);
    }
  }

  async function joinRoom(): Promise<void> {
    if (get(joining)) return;
    const normalized = normalizeCode(get(code));
    if (normalized.length < 3) {
      joinError.set(get(tr)('welcome.joinError'));
      return;
    }
    joining.set(true);
    joinError.set('');
    try {
      const room = await deps.api.resolveRoom(normalized);
      deps.navigation.assign(`/r/${room.code}`);
    } catch (err) {
      joinError.set(errorText(get(tr), errorCodeOf(err, 'invalid_input')));
      joining.set(false);
    }
  }

  return { creating, createError, code, joinError, joining, createRoom, joinRoom };
}
