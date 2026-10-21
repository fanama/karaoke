import { mount } from 'svelte';
import App from './App.svelte';
import { initI18n } from './application/i18n';
import { createApiAdapter } from './adapters/http_api_adapter';
import { connectRoom } from './adapters/ws_socket_adapter';
import { createYouTubePlayer } from './adapters/youtube_iframe_adapter';
import {
  clipboardAdapter,
  clockAdapter,
  fullscreenAdapter,
  navigationAdapter,
  storageAdapter
} from './adapters/browser_adapters';
import type { AppDependencies } from './ports/context';
import './styles/tokens.css';
import './styles/base.css';
import './styles/components.css';

initI18n(storageAdapter);

const deps: AppDependencies = {
  api: createApiAdapter(),
  connectSocket: connectRoom,
  createPlayer: createYouTubePlayer,
  clipboard: clipboardAdapter,
  fullscreen: fullscreenAdapter,
  storage: storageAdapter,
  navigation: navigationAdapter,
  clock: clockAdapter
};

const app = mount(App, { target: document.getElementById('app')!, props: { deps } });

export default app;
