<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { RoomView } from '../../domain/models';
  import type { ConnectionStatus } from '../../ports/socket_port';
  import ConnectionDot from '../common/ConnectionDot.svelte';
  import LanguageToggle from '../common/LanguageToggle.svelte';

  export let room: RoomView;
  export let status: ConnectionStatus;
  export let isHost: boolean;
  export let tab: 'search' | 'queue';
  export let queueCount: number;
  export let onUnlock: () => void;
  export let onTab: (value: 'search' | 'queue') => void;
</script>

<header class="guest-header">
  <div class="topline">
    <a class="mini-brand" href="/">
      <span class="dot"></span>
      {$tr('app.name')}
    </a>
    <div class="row">
      <ConnectionDot {status} compact />
      <LanguageToggle />
      {#if !isHost}
        <button class="lock-btn" type="button" on:click={onUnlock} aria-label={$tr('room.hostControls')}>
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="5" y="11" width="14" height="9" rx="2"/>
            <path d="M8 11V8a4 4 0 0 1 8 0v3"/>
          </svg>
        </button>
      {:else}
        <span class="badge badge-cyan">{$tr('room.host')}</span>
      {/if}
    </div>
  </div>
  <div class="room-head">
    <div>
      <p class="label">{$tr('room.code')}</p>
      <h1 class="room-code">{room.code.toUpperCase()}</h1>
    </div>
    <div class="room-stats">
      <span class="badge">{$tr('room.guests', { count: room.guestCount, max: room.guestCapacity })}</span>
      {#if room.manualOrder}<span class="badge badge-cyan">{$tr('room.manualOrder')}</span>{/if}
    </div>
  </div>
  <div class="tabs" role="tablist">
    <button
      type="button"
      role="tab"
      aria-selected={tab === 'search'}
      class:active={tab === 'search'}
      on:click={() => onTab('search')}
    >
      {$tr('room.tabSearch')}
    </button>
    <button
      type="button"
      role="tab"
      aria-selected={tab === 'queue'}
      class:active={tab === 'queue'}
      on:click={() => onTab('queue')}
    >
      {$tr('room.tabQueue')}
      <span class="count">{queueCount}</span>
    </button>
  </div>
</header>

<style>
  .guest-header {
    position: sticky;
    top: 0;
    z-index: 30;
    padding: 0.75rem 1rem 0;
    background: linear-gradient(to bottom, rgb(8 5 15 / 0.98) 70%, rgb(8 5 15 / 0.86));
    backdrop-filter: blur(12px);
    border-bottom: 1px solid var(--border);
  }

  .topline {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.8rem;
    margin-bottom: 0.8rem;
  }

  .mini-brand {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    color: var(--text);
    text-decoration: none;
    font-weight: 800;
    font-size: 0.86rem;
    letter-spacing: -0.02em;
  }

  .mini-brand .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--magenta);
    box-shadow: 0 0 12px var(--magenta);
  }

  .lock-btn {
    display: grid;
    place-items: center;
    width: 34px;
    height: 34px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--muted);
  }

  .lock-btn:hover {
    color: var(--text);
    border-color: var(--cyan);
  }

  .room-head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 0.8rem;
  }

  .room-code {
    font-family: var(--mono);
    font-size: clamp(1.6rem, 8vw, 2.2rem);
    letter-spacing: 0.08em;
    line-height: 1.1;
    color: var(--text);
  }

  .room-stats {
    display: flex;
    gap: 0.35rem;
    flex-wrap: wrap;
    justify-content: flex-end;
    padding-bottom: 0.2rem;
  }

  .tabs {
    display: flex;
    gap: 0.4rem;
  }

  .tabs button {
    flex: 1;
    position: relative;
    min-height: 44px;
    border: 0;
    border-bottom: 2px solid transparent;
    background: transparent;
    color: var(--dim);
    font-weight: 800;
    font-size: 0.9rem;
  }

  .tabs button.active {
    color: var(--text);
    border-color: var(--magenta);
  }

  .count {
    display: inline-grid;
    place-items: center;
    min-width: 20px;
    height: 20px;
    padding: 0 5px;
    border-radius: 999px;
    background: var(--surface-3);
    font-size: 0.7rem;
    margin-left: 4px;
  }

  @media (min-width: 680px) {
    .guest-header {
      padding-left: 1.5rem;
      padding-right: 1.5rem;
    }
  }
</style>
