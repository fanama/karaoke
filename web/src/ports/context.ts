import { getContext, setContext } from 'svelte';
import type { RoomApiPort } from './api_port';
import type { SocketFactory } from './socket_port';
import type { PlayerFactory } from './player_port';
import type { ClipboardPort, ClockPort, FullscreenPort, NavigationPort, StoragePort } from './browser_port';

/** Everything a view needs, injected once by main.ts. */
export interface AppDependencies {
  api: RoomApiPort;
  connectSocket: SocketFactory;
  createPlayer: PlayerFactory;
  clipboard: ClipboardPort;
  fullscreen: FullscreenPort;
  storage: StoragePort;
  navigation: NavigationPort;
  clock: ClockPort;
}

const KEY = Symbol('fana.deps');

export function setAppContext(deps: AppDependencies): AppDependencies {
  setContext(KEY, deps);
  return deps;
}

export function getAppContext(): AppDependencies {
  const deps = getContext<AppDependencies | undefined>(KEY);
  if (!deps) throw new Error('App dependencies were not provided');
  return deps;
}
