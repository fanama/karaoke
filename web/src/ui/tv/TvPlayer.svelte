<script lang="ts">
  import { createEventDispatcher, onDestroy, onMount } from 'svelte';
  import { tr } from '../../application/i18n';
  import Icon from '../common/Icon.svelte';
  import type { Status } from '../../domain/models';
  import type { PlayerFailure, PlayerProgress, PlayerTelemetry, VideoPlayerPort } from '../../ports/player_port';
  import { getAppContext } from '../../ports/context';

  export let videoId: string | null = null;
  export let entryId = '';
  export let generation = 0;
  export let status: Status = 'idle';
  export let elapsed = 0;
  export let muted = false;
  export let blocked = false;

  const deps = getAppContext();
  const dispatch = createEventDispatcher<{
    ended: PlayerTelemetry;
    error: PlayerFailure;
    progress: PlayerProgress;
    blocked: boolean;
    soundblocked: boolean;
  }>();

  let container: HTMLDivElement;
  let player: VideoPlayerPort | null = null;

  onMount(() => {
    player = deps.createPlayer(container, {
      ended: (detail) => dispatch('ended', detail),
      error: (detail) => dispatch('error', detail),
      progress: (detail) => dispatch('progress', detail),
      blocked: (value) => dispatch('blocked', value),
      soundblocked: (value) => dispatch('soundblocked', value)
    });
  });

  onDestroy(() => player?.destroy());

  $: player?.update({ videoId, entryId, generation, status, elapsed });
  $: player?.setMuted(muted);

  export function forcePlay() {
    player?.forcePlay();
  }

  export function setMuted(value: boolean) {
    player?.setMuted(value);
  }
</script>

<div class="player-wrap">
  <div class="player" bind:this={container}></div>
  {#if blocked}
    <button class="blocked-overlay" type="button" on:click={forcePlay}>
      <span class="tap-ring" aria-hidden="true"><Icon name="play" size={28} /></span>
      <strong>{$tr('tv.tapResume')}</strong>
      <small>{$tr('tv.tapResumeBody')}</small>
    </button>
  {/if}
</div>

<style>
  .player-wrap {
    position: absolute;
    inset: 0;
    background: var(--bg);
  }

  .player,
  .player :global(iframe) {
    width: 100%;
    height: 100%;
    border: 0;
  }

  .player {
    pointer-events: auto;
  }

  .blocked-overlay {
    position: absolute;
    inset: 0;
    z-index: 5;
    display: grid;
    place-content: center;
    justify-items: center;
    gap: var(--space-2);
    border: 0;
    background: rgb(11 11 12 / 0.86);
    color: var(--text);
    text-align: center;
    padding: var(--space-4);
  }

  .tap-ring {
    display: grid;
    place-items: center;
    width: 84px;
    height: 84px;
    margin-bottom: var(--space-2);
    border: 2px solid var(--accent);
    color: var(--accent);
  }

  .blocked-overlay strong {
    font-size: var(--text-lg);
    text-transform: uppercase;
  }

  .blocked-overlay small {
    max-width: 42ch;
    color: var(--muted);
    font-size: var(--text-xs);
    line-height: 1.5;
  }
</style>
