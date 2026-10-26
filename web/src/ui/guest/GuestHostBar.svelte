<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { Snapshot } from '../../domain/models';
  import Icon from '../common/Icon.svelte';

  export let snap: Snapshot;
  export let busy: boolean;
  export let onPlayback: (action: 'play' | 'pause' | 'next' | 'cancelCountdown') => void;
  export let onToggleLock: () => void;
  export let onManage: () => void;
</script>

<div class="host-bar">
  <div class="host-playback">
    {#if snap.nowPlaying?.status === 'playing'}
      <button
        class="hbtn"
        type="button"
        on:click={() => onPlayback('pause')}
        disabled={busy}
        aria-label={$tr('tv.pause')}
      >
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <rect x="6" y="5" width="4" height="14" /><rect x="14" y="5" width="4" height="14" />
        </svg>
      </button>
    {:else}
      <button
        class="hbtn"
        type="button"
        on:click={() => onPlayback('play')}
        disabled={busy}
        aria-label={$tr('tv.play')}
      >
        <Icon name="play" size={20} />
      </button>
    {/if}
    <button
      class="hbtn"
      type="button"
      on:click={() => onPlayback('next')}
      disabled={busy}
      aria-label={$tr('tv.next')}
    >
      <Icon name="skip" size={20} />
    </button>
    {#if snap.nowPlaying?.status === 'countdown'}
      <button
        class="hbtn"
        type="button"
        on:click={() => onPlayback('cancelCountdown')}
        disabled={busy}
        aria-label={$tr('tv.cancel')}
        title={$tr('tv.cancel')}
      >
        <Icon name="close" size={18} />
      </button>
    {/if}
  </div>

  <span class="host-mode-label">{$tr('room.hostMode')}</span>

  <div class="host-playback">
    <button
      class="hbtn"
      type="button"
      on:click={onToggleLock}
      aria-label={$tr(snap.room.queueLocked ? 'room.unlock' : 'room.lock')}
    >
      <Icon name="lock" size={18} />
    </button>
    <button class="btn btn-secondary btn-sm" type="button" on:click={onManage}>
      {$tr('room.manage')}
    </button>
  </div>
</div>

<style>
  .host-bar {
    position: fixed;
    left: 50%;
    bottom: 0;
    transform: translateX(-50%);
    z-index: 40;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    width: min(720px, 100%);
    padding: var(--space-3) var(--space-4) calc(var(--space-3) + env(safe-area-inset-bottom));
    background: var(--surface);
  }

  .host-playback {
    display: flex;
    gap: var(--space-1);
  }

  .hbtn {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    padding: 0;
    border: 0;
    border-radius: var(--radius-sm);
    background: var(--surface-2);
    color: var(--text);
  }

  .hbtn:hover:not(:disabled) {
    background: var(--surface-3);
    color: var(--accent);
  }

  .hbtn:disabled {
    opacity: 0.45;
  }

  .host-mode-label {
    color: var(--accent);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }
</style>
