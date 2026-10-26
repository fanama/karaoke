<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { ConnectionStatus } from '../../ports/socket_port';
  import ConnectionDot from '../common/ConnectionDot.svelte';
  import Icon from '../common/Icon.svelte';
  import LanguageToggle from '../common/LanguageToggle.svelte';
  import TvJoinPanel from './TvJoinPanel.svelte';

  export let code: string;
  export let pin: string;
  export let guestCount: number;
  export let status: ConnectionStatus;
  export let onCopy: () => void;
</script>

<header class="tv-top">
  <TvJoinPanel {code} {pin} />

  <div class="room-mark">
    <p class="label">{$tr('room.code')}</p>
    <h1 class="tv-code">{code.toUpperCase()}</h1>
    <div class="row">
      <ConnectionDot {status} />
      <span class="tv-stat">{$tr('tv.guestCount', { count: guestCount })}</span>
    </div>
  </div>

  <div class="tv-tools">
    <LanguageToggle />
    <button class="tool-btn" type="button" on:click={onCopy} title={$tr('room.copyTvLink')}>
      <Icon name="link" size={18} />
    </button>
    <a class="tool-btn" href={`/r/${code}`} title={$tr('room.viewGuest')}>
      <Icon name="phone" size={18} />
    </a>
  </div>
</header>

<style>
  .tv-top {
    position: relative;
    z-index: 10;
    display: flex;
    align-items: flex-start;
    gap: var(--space-5);
    padding: var(--space-4) var(--space-5);
    background: var(--bg);
    pointer-events: none;
  }

  .tv-top > * {
    pointer-events: auto;
  }

  .room-mark {
    display: grid;
    gap: var(--space-1);
    min-width: 0;
  }

  .tv-code {
    color: var(--text);
    font-family: var(--mono);
    font-size: var(--text-2xl);
    font-weight: 800;
    letter-spacing: 0.12em;
    line-height: 1;
    text-transform: uppercase;
  }

  .tv-stat {
    color: var(--muted);
    font-size: var(--text-2xs);
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .tv-tools {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-left: auto;
  }

  .tool-btn {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    padding: 0;
    border: 0;
    border-radius: var(--radius-sm);
    background: var(--surface);
    color: var(--text);
    text-decoration: none;
  }

  .tool-btn:hover {
    background: var(--surface-3);
    color: var(--accent);
  }

  @media (max-width: 900px) {
    .tv-top {
      gap: var(--space-3);
      padding: var(--space-3);
    }

    .tv-code {
      font-size: var(--text-lg);
    }
  }

  @media (max-width: 600px) {
    .tv-tools {
      gap: var(--space-1);
    }
  }
</style>
