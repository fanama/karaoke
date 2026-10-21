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
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 15;
    display: grid;
    grid-template-columns: minmax(260px, 1.2fr) auto minmax(220px, 1fr);
    align-items: end;
    gap: 1.5rem;
    padding: 1.4rem 1.7rem 1.5rem;
    background: linear-gradient(to top, rgb(5 2 10 / 0.96), rgb(5 2 10 / 0.6) 60%, transparent);
    transition: opacity 0.25s ease, transform 0.25s ease;
  }

  .tv-bottom.hidden {
    opacity: 0;
    transform: translateY(12px);
    pointer-events: none;
  }

  .now-summary {
    display: flex;
    gap: 0.8rem;
    align-items: center;
    min-width: 0;
  }

  .now-thumb {
    width: 116px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 10px;
    border: 1px solid var(--border);
    flex: none;
    background: var(--surface-3);
  }

  .now-meta {
    min-width: 0;
    flex: 1;
  }

  .now-title {
    font-size: 1.08rem;
    font-weight: 800;
    line-height: 1.25;
    margin-bottom: 0.2rem;
  }

  .timeline {
    height: 4px;
    border-radius: 999px;
    background: rgb(255 255 255 / 0.14);
    overflow: hidden;
    margin-top: 0.4rem;
  }

  .timeline span {
    display: block;
    height: 100%;
    background: linear-gradient(90deg, var(--magenta), var(--cyan));
  }

  .playback-controls {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.4rem;
    border-radius: 999px;
    background: rgb(18 10 31 / 0.8);
    border: 1px solid var(--border);
    backdrop-filter: blur(10px);
  }

  .ctrl {
    position: relative;
    display: grid;
    place-items: center;
    width: 48px;
    height: 48px;
    padding: 0;
    border: 0;
    border-radius: 50%;
    background: transparent;
    color: var(--text);
  }

  .ctrl:hover {
    background: rgb(255 255 255 / 0.07);
  }

  .ctrl.primary {
    width: 58px;
    height: 58px;
    background: linear-gradient(135deg, var(--magenta), #ff63b4);
    color: #1a0010;
    box-shadow: var(--shadow-neon);
  }

  .queue-badge {
    position: absolute;
    top: -3px;
    right: -4px;
    min-width: 18px;
    height: 18px;
    padding: 0 4px;
    display: grid;
    place-items: center;
    border-radius: 999px;
    background: var(--cyan);
    color: #001014;
    font-size: 0.62rem;
    font-weight: 900;
  }

  @media (max-width: 900px) {
    .tv-bottom {
      grid-template-columns: 1fr auto;
      gap: 0.8rem;
    }
  }

  @media (max-width: 600px) {
    .now-thumb {
      width: 74px;
    }
    .now-title {
      font-size: 0.85rem;
    }
    .tv-bottom {
      padding: 0.7rem;
    }
    .playback-controls {
      gap: 0;
      padding: 0.2rem;
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
