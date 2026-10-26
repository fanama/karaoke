<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { QueueItem, Snapshot } from '../../domain/models';
  import { computeProgress, formatDuration } from '../../domain/playback';
  import TvNextSummary from './TvNextSummary.svelte';

  export let snap: Snapshot;
  export let queue: QueueItem[];
  export let elapsedNow: number;
  export let muted: boolean;
  export let busy: boolean;
  export let hidden: boolean;
  export let onPlayback: (action: 'play' | 'pause' | 'next') => void;
  export let onToggleMute: () => void;
  export let onFullscreen: () => void;
  export let onToggleQueue: () => void;
</script>

<div class="tv-bottom" class:hidden>
  <div class="now-summary">
    {#if snap.nowPlaying}
      <img class="now-thumb" src={snap.nowPlaying.thumbnailUrl} alt="" />
      <div class="now-meta">
        <p class="now-title clamp-2">{snap.nowPlaying.title}</p>
        <p class="tiny muted truncate">
          {$tr('tv.nowSinging')}: <strong>{snap.nowPlaying.nickname}</strong>
          {#if snap.nowPlaying.channel} · {snap.nowPlaying.channel}{/if}
        </p>
        <div class="timeline">
          <span style={`width:${computeProgress(elapsedNow, snap.nowPlaying.durationSeconds)}%`}></span>
        </div>
        <div class="spread tiny dim">
          <span>{formatDuration(elapsedNow)}</span>
          <span>{formatDuration(snap.nowPlaying.durationSeconds)}</span>
        </div>
      </div>
    {:else}
      <div class="now-meta">
        <p class="now-title">{$tr('tv.noSong')}</p>
        <p class="tiny muted">{$tr('tv.addSongs')}</p>
      </div>
    {/if}
  </div>

  <div class="playback-controls">
    {#if snap.nowPlaying?.status === 'playing'}
      <button class="ctrl" type="button" on:click={() => onPlayback('pause')} disabled={busy} title={$tr('tv.pause')}>
        <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor"><rect x="6" y="5" width="4" height="14" rx="1"/><rect x="14" y="5" width="4" height="14" rx="1"/></svg>
      </button>
    {:else}
      <button class="ctrl primary" type="button" on:click={() => onPlayback('play')} disabled={busy} title={$tr('tv.play')}>
        <svg viewBox="0 0 24 24" width="26" height="26" fill="currentColor"><path d="M7 5v14l12-7z"/></svg>
      </button>
    {/if}
    <button class="ctrl" type="button" on:click={() => onPlayback('next')} disabled={busy} title={$tr('tv.next')}>
      <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor"><path d="M6 5v14l9-7z"/><rect x="16" y="5" width="2" height="14" rx="1"/></svg>
    </button>
    <button class="ctrl" type="button" on:click={onToggleMute} title={$tr(muted ? 'tv.unmute' : 'tv.mute')}>
      <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M4 9v6h4l5 4V5L8 9z"/>
        {#if muted}<path d="m17 9 4 6M21 9l-4 6"/>{:else}<path d="M17 8a6 6 0 0 1 0 8"/>{/if}
      </svg>
    </button>
    <button class="ctrl" type="button" on:click={onFullscreen} title={$tr('tv.fullscreen')}>
      <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5"/></svg>
    </button>
    <button class="ctrl queue-toggle" type="button" on:click={onToggleQueue} title={$tr('tv.openQueue')}>
      <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01"/></svg>
      <span class="queue-badge">{snap.queue.length}</span>
    </button>
  </div>

  <TvNextSummary next={queue[0]} />
</div>

<style>
  .tv-bottom {
    position: relative;
    z-index: 15;
    display: grid;
    grid-template-columns: minmax(260px, 1.2fr) auto minmax(220px, 1fr);
    align-items: center;
    gap: var(--space-5);
    padding: var(--space-4) var(--space-5);
    background: var(--bg);
    transition: opacity 0.2s ease, transform 0.2s ease;
  }

  .tv-bottom.hidden {
    opacity: 0;
    transform: translateY(12px);
    pointer-events: none;
  }

  .now-summary {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-width: 0;
  }

  .now-thumb {
    width: 116px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    flex: none;
    background: var(--surface-2);
    border: 1px solid var(--border);
  }

  .now-meta {
    min-width: 0;
    flex: 1;
  }

  .now-title {
    margin-bottom: 2px;
    font-size: var(--text-sm);
    font-weight: 800;
    line-height: 1.25;
  }

  .timeline {
    height: 4px;
    margin-top: var(--space-2);
    background: var(--surface-3);
    overflow: hidden;
  }

  .timeline span {
    display: block;
    height: 100%;
    background: var(--accent);
  }

  .playback-controls {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1);
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
  }

  .ctrl {
    position: relative;
    display: grid;
    place-items: center;
    width: 48px;
    height: 48px;
    padding: 0;
    border: 0;
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--text);
  }

  .ctrl:hover {
    background: var(--surface-2);
  }

  .ctrl.primary {
    width: 54px;
    height: 54px;
    background: var(--accent);
    color: var(--accent-ink);
  }

  .queue-badge {
    position: absolute;
    top: -4px;
    right: -4px;
    display: grid;
    place-items: center;
    min-width: 20px;
    height: 20px;
    padding: 0 var(--space-1);
    background: var(--accent);
    color: var(--accent-ink);
    border-radius: var(--radius-sm);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    font-weight: 800;
  }

  @media (max-width: 900px) {
    .tv-bottom {
      grid-template-columns: 1fr auto;
      gap: var(--space-3);
    }
  }

  @media (max-width: 600px) {
    .now-thumb {
      width: 74px;
    }

    .tv-bottom {
      padding: var(--space-2);
    }

    .playback-controls {
      padding: var(--space-1);
    }

    .ctrl {
      width: 40px;
      height: 40px;
    }

    .ctrl.primary {
      width: 46px;
      height: 46px;
    }
  }
</style>
