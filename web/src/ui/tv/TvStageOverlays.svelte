<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { Snapshot } from '../../domain/models';
  import Icon from '../common/Icon.svelte';

  export let snap: Snapshot;
  export let countdownLeft: number;
  export let onPlay: () => void;
  export let onCancelCountdown: () => void;
</script>

{#if !snap.room.tvStarted}
  <div class="stage-overlay">
    <div class="overlay-card paper">
      <span class="overlay-mark" aria-hidden="true"><Icon name="microphone" size={26} /></span>
      <h2>{$tr('tv.start')}</h2>
      <p>{$tr('tv.startBody')}</p>
      <button class="btn btn-primary btn-lg" type="button" on:click={onPlay}>
        {$tr('tv.start')}
      </button>
    </div>
  </div>
{:else if !snap.nowPlaying}
  <div class="stage-overlay">
    <div class="overlay-card paper">
      <span class="overlay-mark" aria-hidden="true"><Icon name="microphone" size={26} /></span>
      <h2>{$tr('tv.idleTitle')}</h2>
      <p>{$tr('tv.idleBody')}</p>
      <span class="badge">{$tr('tv.addSongs')}</span>
    </div>
  </div>
{:else if snap.nowPlaying.status === 'countdown'}
  <div class="stage-overlay">
    <div class="overlay-card paper countdown-card">
      <p class="countdown-label">{$tr('tv.countdownNext')}</p>
      <h2 class="clamp-2">{snap.nowPlaying.title}</h2>
      <p class="countdown-singer">
        {$tr('tv.nowSinging')}: <strong>{snap.nowPlaying.nickname}</strong>
      </p>
      <div class="countdown-number">{Math.max(1, countdownLeft)}</div>
      <button class="btn btn-secondary" type="button" on:click={onCancelCountdown}>
        {$tr('tv.cancel')}
      </button>
    </div>
  </div>
{/if}

<style>
  .stage-overlay {
    position: absolute;
    inset: 0;
    z-index: 8;
    display: grid;
    place-items: center;
    padding: var(--space-6);
    pointer-events: none;
  }

  .stage-overlay > * {
    pointer-events: auto;
  }

  .overlay-card {
    display: grid;
    justify-items: center;
    gap: var(--space-3);
    max-width: 520px;
    padding: var(--space-6);
    text-align: center;
  }

  .overlay-mark {
    display: grid;
    place-items: center;
    width: 56px;
    height: 56px;
    border: 1px solid var(--ink);
    color: var(--ink);
  }

  .overlay-card h2 {
    font-size: var(--text-xl);
    font-weight: 800;
    letter-spacing: -0.03em;
    line-height: 1.05;
    text-transform: uppercase;
    text-wrap: balance;
  }

  .overlay-card p {
    max-width: 38ch;
    color: var(--ink-muted);
    font-size: var(--text-sm);
    line-height: 1.55;
  }

  .overlay-card .badge {
    border-color: var(--ink-muted);
    color: var(--ink-muted);
  }

  .overlay-card .btn {
    margin-top: var(--space-2);
  }

  .countdown-label {
    color: var(--ink-muted) !important;
    font-family: var(--mono);
    font-size: var(--text-2xs);
    font-weight: 700;
    letter-spacing: 0.2em;
    text-transform: uppercase;
  }

  .countdown-singer {
    color: var(--ink-muted) !important;
    font-size: var(--text-xs) !important;
  }

  .countdown-number {
    margin: var(--space-1) 0 var(--space-2);
    color: var(--ink);
    font-family: var(--mono);
    font-size: clamp(4rem, 9vw, 7rem);
    line-height: 0.9;
  }
</style>
