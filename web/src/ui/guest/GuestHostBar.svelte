<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { Snapshot } from '../../domain/models';

  export let snap: Snapshot;
  export let busy: boolean;
  export let onPlayback: (action: 'play' | 'pause' | 'next' | 'cancelCountdown') => void;
  export let onToggleLock: () => void;
  export let onManage: () => void;
</script>

<div class="host-bar">
  <div class="host-playback">
    {#if snap.nowPlaying?.status === 'playing'}
      <button class="hbtn" type="button" on:click={() => onPlayback('pause')} disabled={busy} aria-label={$tr('tv.pause')}>
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><rect x="6" y="5" width="4" height="14" rx="1"/><rect x="14" y="5" width="4" height="14" rx="1"/></svg>
      </button>
    {:else}
      <button class="hbtn" type="button" on:click={() => onPlayback('play')} disabled={busy} aria-label={$tr('tv.play')}>
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M7 5v14l12-7z"/></svg>
      </button>
    {/if}
    <button class="hbtn" type="button" on:click={() => onPlayback('next')} disabled={busy} aria-label={$tr('tv.next')}>
      <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M6 5v14l9-7z"/><rect x="16" y="5" width="2" height="14" rx="1"/></svg>
    </button>
    {#if snap.nowPlaying?.status === 'countdown'}
      <button class="hbtn" type="button" on:click={() => onPlayback('cancelCountdown')} disabled={busy} aria-label={$tr('tv.cancel')} title={$tr('tv.cancel')}>
        ✕
      </button>
    {/if}
  </div>
  <span class="host-mode-label">{$tr('room.hostMode')}</span>
  <div class="host-playback">
    <button class="hbtn" type="button" on:click={onToggleLock} aria-label={$tr(snap.room.queueLocked ? 'room.unlock' : 'room.lock')}>
      <svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="2">
        <rect x="5" y="11" width="14" height="9" rx="2"/>
        {#if snap.room.queueLocked}<path d="M8 11V8a4 4 0 0 1 7.5-1.9"/>{:else}<path d="M8 11V8a4 4 0 0 1 8 0v3"/>{/if}
      </svg>
    </button>
    <button class="btn btn-secondary btn-sm" type="button" on:click={onManage}>{$tr('room.manage')}</button>
  </div>
</div>

<style>
  .host-bar {
    position: fixed;
    left: 50%;
    bottom: 0;
    transform: translateX(-50%);
    width: min(720px, 100%);
    z-index: 40;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.6rem;
    padding: 0.6rem 1rem calc(0.6rem + env(safe-area-inset-bottom));
    background: rgb(18 10 31 / 0.97);
    border-top: 1px solid var(--border-bright);
    backdrop-filter: blur(14px);
  }

  .host-playback {
    display: flex;
    gap: 0.35rem;
  }

  .hbtn {
    display: grid;
    place-items: center;
    width: 42px;
    height: 42px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: var(--surface-2);
    color: var(--text);
  }

  .hbtn:hover {
    border-color: var(--magenta);
  }

  .host-mode-label {
    font-size: 0.66rem;
    font-weight: 800;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--magenta);
  }
</style>
