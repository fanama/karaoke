import { get, writable } from 'svelte/store';
import { errorCodeOf } from '../domain/errors';
import type { AdminSession, AdminSong } from '../domain/models';
import { errorText, tr } from './i18n';
import type { AppDependencies } from '../ports/context';

/** SongPageSize is how many titles the admin list shows per page. */
export const SongPageSize = 25;

/** FilterDelay is how long typing waits before the filtered page is fetched. */
export const FilterDelay = 300;

function createAdminState() {
  return {
    session: writable<AdminSession | null>(null),
    token: writable(''),
    format: writable<'sqlite' | 'sql'>('sqlite'),
    busy: writable(false),
    error: writable(''),
    notice: writable(''),
    songs: writable<AdminSong[]>([]),
    songsTotal: writable(0),
    songsOffset: writable(0),
    songsQuery: writable(''),
    loadingSongs: writable(false),
    songsError: writable('')
  };
}

type AdminState = ReturnType<typeof createAdminState>;

/** Admin page state: token login, library transfer and the paged song list. */
export function createAdminController(deps: AppDependencies) {
  const state = createAdminState();
  const actions = new AdminActions(state, deps);
  return {
    ...state,
    load: () => actions.load(),
    login: () => actions.login(),
    logout: () => actions.logout(),
    importLibrary: (file: File) => actions.importLibrary(file),
    filterSongs: (value: string) => actions.filterSongs(value),
    previousPage: () => actions.goToPage(get(state.songsOffset) - SongPageSize),
    nextPage: () => actions.goToPage(get(state.songsOffset) + SongPageSize),
    reloadSongs: () => actions.loadSongs(get(state.songsOffset))
  };
}

class AdminActions {
  private filterTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(
    private readonly state: AdminState,
    private readonly deps: AppDependencies
  ) {}

  async load(): Promise<void> {
    this.state.error.set('');
    try {
      const session = await this.deps.api.adminSession();
      this.state.session.set(session);
      if (session.authenticated && session.library) {
        await this.loadSongs(0);
      }
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
      await this.loadSongs(0);
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
    this.state.songs.set([]);
    this.state.songsTotal.set(0);
    this.state.songsOffset.set(0);
    this.state.songsQuery.set('');
    await this.load();
  }

  async importLibrary(file: File): Promise<void> {
    if (get(this.state.busy)) return;
    this.state.busy.set(true);
    this.state.error.set('');
    this.state.notice.set('');
    try {
      const res = await this.deps.api.adminImport(file, get(this.state.format));
      this.state.notice.set(get(tr)('admin.imported', { count: res.imported, total: res.songs }));
      await this.load();
    } catch (err) {
      this.state.error.set(errorText(get(tr), errorCodeOf(err, 'library_invalid')));
    } finally {
      this.state.busy.set(false);
    }
  }

  /** filterSongs debounces so a typed filter costs one request, not one per key. */
  filterSongs(value: string): void {
    this.state.songsQuery.set(value);
    if (this.filterTimer) clearTimeout(this.filterTimer);
    this.filterTimer = setTimeout(() => void this.loadSongs(0), FilterDelay);
  }

  goToPage(offset: number): Promise<void> {
    return this.loadSongs(offset);
  }

  async loadSongs(offset: number): Promise<void> {
    if (get(this.state.loadingSongs)) return;
    this.state.loadingSongs.set(true);
    this.state.songsError.set('');
    try {
      const page = await this.deps.api.adminSongs(
        Math.max(0, offset),
        SongPageSize,
        get(this.state.songsQuery)
      );
      this.state.songs.set(page.songs);
      this.state.songsTotal.set(page.total);
      this.state.songsOffset.set(page.offset);
    } catch (err) {
      this.state.songsError.set(errorText(get(tr), errorCodeOf(err, 'internal_error')));
    } finally {
      this.state.loadingSongs.set(false);
    }
  }
}
