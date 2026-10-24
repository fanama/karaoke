<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { ConnectionStatus } from '../../ports/socket_port';
  import ConnectionDot from '../common/ConnectionDot.svelte';
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
      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7 0l2-2a5 5 0 0 0-7-7l-1 1"/><path d="M14 11a5 5 0 0 0-7 0l-2 2a5 5 0 0 0 7 7l1-1"/></svg>
    </button>
    <a class="tool-btn" href={`/r/${code}`} title={$tr('room.viewGuest')}>
      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><rect x="7" y="2" width="10" height="20" rx="2"/><path d="M11 18h2"/></svg>
    </a>
  </div>
</header>

<style>
  .tv-top {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    z-index: 10;
    display: flex;
    align-items: flex-start;
    gap: 1.2rem;
    padding: 1.4rem 1.7rem;
    pointer-events: none;
  }

  .tv-top > * {
    pointer-events: auto;
  }

  .room-mark {
    display: grid;
    gap: 0.25rem;
    min-width: 0;
    text-shadow: 0 2px 20px rgb(0 0 0 / 0.7);
  }

  .tv-code {
    font-family: var(--mono);
    font-size: clamp(1.7rem, 2.8vw, 2.7rem);
    letter-spacing: 0.12em;
    line-height: 1;
    color: #fff;
  }

  .tv-stat {
    font-size: 0.78rem;
    color: var(--muted);
    font-weight: 700;
  }

  .tv-tools {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin-left: auto;
  }

  .tool-btn {
    position: relative;
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: rgb(8 5 15 / 0.65);
    color: var(--text);
    text-decoration: none;
  }

  .tool-btn:hover {
    border-color: var(--cyan);
  }

  @media (max-width: 600px) {
    .tv-top {
      padding: 0.8rem;
    }
    .tv-code {
      font-size: 1.3rem;
    }
  }
</style>
