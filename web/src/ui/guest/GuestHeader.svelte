<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { RoomView } from '../../domain/models';
  import type { ConnectionStatus } from '../../ports/socket_port';
  import ConnectionDot from '../common/ConnectionDot.svelte';
  import Icon from '../common/Icon.svelte';
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
      <span class="mark" aria-hidden="true"></span>
      {$tr('app.name')}
    </a>
    <div class="row">
      <ConnectionDot {status} compact />
      <LanguageToggle />
      {#if !isHost}
        <button
          class="btn btn-icon"
          type="button"
          data-testid="unlock-host"
          on:click={onUnlock}
          aria-label={$tr('room.hostControls')}
        >
          <Icon name="lock" size={18} />
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
    padding: var(--space-4);
    background: var(--bg);
  }

  .topline {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
  }

  .mini-brand {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--text);
    text-decoration: none;
    font-size: var(--text-2xs);
    font-weight: 800;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .mark {
    width: 8px;
    height: 8px;
    background: var(--accent);
  }

  .guest-header :global(.btn-icon) {
    width: 36px;
    height: 36px;
    min-height: 36px;
  }

  .room-head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--space-4);
    margin-bottom: var(--space-4);
  }

  .room-code {
    font-family: var(--mono);
    font-size: var(--text-xl);
    font-weight: 800;
    letter-spacing: 0.08em;
    line-height: 1.1;
    text-transform: uppercase;
  }

  .room-stats {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: var(--space-1);
    padding-bottom: 2px;
  }

  .tabs {
    display: flex;
    gap: var(--space-2);
  }

  .tabs button {
    flex: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-1);
    min-height: 44px;
    border: 0;
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--dim);
    font-size: var(--text-xs);
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .tabs button.active {
    background: var(--surface-2);
    color: var(--text);
  }

  .count {
    display: inline-grid;
    place-items: center;
    min-width: 20px;
    height: 20px;
    padding: 0 var(--space-1);
    border-radius: var(--radius-sm);
    background: var(--surface-3);
    color: var(--muted);
    font-family: var(--mono);
    font-size: var(--text-2xs);
  }

  @media (min-width: 680px) {
    .guest-header {
      padding-inline: var(--space-5);
    }
  }
</style>
