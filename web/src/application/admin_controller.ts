import { get, writable } from 'svelte/store';
import { errorCodeOf } from '../domain/errors';
import type { AdminSession } from '../domain/models';
import { errorText, tr } from './i18n';
import type { AppDependencies } from '../ports/context';

function createAdminState() {
  return {
    session: writable<AdminSession | null>(null),
    token: writable(''),
    busy: writable(false),
    error: writable(''),
    notice: writable('')
  };
}

type AdminState = ReturnType<typeof createAdminState>;

/** Admin page state: token login and SQLite library transfer. */
export function createAdminController(deps: AppDependencies) {
  const state = createAdminState();
  const actions = new AdminActions(state, deps);
  return {
    ...state,
    load: () => actions.load(),
    login: () => actions.login(),
    logout: () => actions.logout(),
    importLibrary: (file: File) => actions.importLibrary(file)
  };
}

class AdminActions {
  constructor(
    private readonly state: AdminState,
    private readonly deps: AppDependencies
  ) {}

  async load(): Promise<void> {
    this.state.error.set('');
    try {
      this.state.session.set(await this.deps.api.adminSession());
    } catch {
      this.state.error.set(get(tr)('admin.loadError'));
    }
  }

  async login(): Promise<void> {
    if (get(this.state.busy)) return;
    this.state.busy.set(true);
    this.state.error.set('');
    this.state.notice.set('');
    try {
      this.state.session.set(await this.deps.api.adminLogin(get(this.state.token).trim()));
      this.state.token.set('');
    } catch (err) {
      this.state.error.set(errorText(get(tr), errorCodeOf(err, 'invalid_token')));
    } finally {
      this.state.busy.set(false);
    }
  }

  async logout(): Promise<void> {
    try {
      await this.deps.api.adminLogout();
    } catch {
      // Reloading the session below reflects the cookie state either way.
    }
    await this.load();
  }

  async importLibrary(file: File): Promise<void> {
    if (get(this.state.busy)) return;
    this.state.busy.set(true);
    this.state.error.set('');
    this.state.notice.set('');
    try {
      const res = await this.deps.api.adminImport(file);
      this.state.notice.set(get(tr)('admin.imported', { count: res.imported, total: res.songs }));
      await this.load();
    } catch (err) {
      this.state.error.set(errorText(get(tr), errorCodeOf(err, 'library_invalid')));
    } finally {
      this.state.busy.set(false);
    }
  }
}
