<script lang="ts">
  import { createEventDispatcher, onDestroy, onMount } from 'svelte';
  import { tr } from '../../application/i18n';
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
      <span class="tap-ring" aria-hidden="true">▶</span>
      <strong>{$tr('tv.tapResume')}</strong>
      <small>{$tr('tv.tapResumeBody')}</small>
    </button>
  {/if}
</div>

<style>
  .player-wrap {
    position: absolute;
    inset: 0;
    background: #000;
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
    gap: 0.5rem;
    border: 0;
    background: rgb(8 5 15 / 0.72);
    backdrop-filter: blur(5px);
    color: var(--text);
    text-align: center;
    padding: 1rem;
  }

  .tap-ring {
    display: grid;
    place-items: center;
    width: 84px;
    height: 84px;
    border-radius: 50%;
    border: 2px solid var(--magenta);
    color: var(--magenta);
    font-size: 1.8rem;
    box-shadow: var(--shadow-neon);
    margin-bottom: 0.5rem;
  }

  .blocked-overlay strong {
    font-size: clamp(1.2rem, 2.4vw, 1.8rem);
  }

  .blocked-overlay small {
    color: var(--muted);
    font-size: 0.9rem;
  }
</style>
