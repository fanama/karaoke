<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { Snapshot } from '../../domain/models';

  export let snap: Snapshot;
  export let countdownLeft: number;
  export let onPlay: () => void;
  export let onCancelCountdown: () => void;
</script>

{#if !snap.room.tvStarted}
  <div class="stage-overlay start-overlay">
    <div class="start-card">
      <div class="start-icon" aria-hidden="true">🎤</div>
      <h2>{$tr('tv.start')}</h2>
      <p>{$tr('tv.startBody')}</p>
      <button class="btn btn-primary btn-lg" type="button" on:click={onPlay}>
        {$tr('tv.start')}
      </button>
    </div>
  </div>
{:else if !snap.nowPlaying}
  <div class="stage-overlay idle-overlay">
    <div class="idle-card">
      <div class="idle-note">♪</div>
      <h2>{$tr('tv.idleTitle')}</h2>
      <p>{$tr('tv.idleBody')}</p>
      <span class="badge badge-cyan">{$tr('tv.addSongs')}</span>
    </div>
  </div>
{:else if snap.nowPlaying.status === 'countdown'}
  <div class="stage-overlay countdown-overlay">
    <div class="countdown-card">
      <p class="countdown-label">{$tr('tv.countdownNext')}</p>
      <h2 class="clamp-2">{snap.nowPlaying.title}</h2>
      <p class="countdown-singer">{$tr('tv.nowSinging')}: <strong>{snap.nowPlaying.nickname}</strong></p>
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
    padding: 2rem;
    pointer-events: none;
  }

  .stage-overlay > * {
    pointer-events: auto;
  }

  .start-card,
  .idle-card,
  .countdown-card {
    display: grid;
    justify-items: center;
    gap: 0.7rem;
    text-align: center;
    padding: 2.2rem;
    border-radius: 24px;
    border: 1px solid var(--border-bright);
    background: rgb(12 7 22 / 0.82);
    backdrop-filter: blur(14px);
    box-shadow: 0 24px 100px rgb(0 0 0 / 0.55);
    max-width: 520px;
  }

  .start-icon,
  .idle-note {
    font-size: 2.4rem;
    line-height: 1;
    margin-bottom: 0.2rem;
  }

  .start-card h2,
  .idle-card h2,
  .countdown-card h2 {
    font-size: clamp(1.5rem, 2.6vw, 2.2rem);
  }

  .start-card p,
  .idle-card p {
    color: var(--muted);
    line-height: 1.55;
    max-width: 38ch;
  }

  .start-card .btn {
    margin-top: 0.6rem;
  }

  .countdown-label {
    font-size: 0.7rem;
    font-weight: 800;
    letter-spacing: 0.2em;
    text-transform: uppercase;
    color: var(--cyan);
  }

  .countdown-singer {
    color: var(--muted);
    font-size: 0.9rem;
  }

  .countdown-number {
    font-family: var(--mono);
    font-size: clamp(4rem, 9vw, 7rem);
    line-height: 0.9;
    color: var(--magenta);
    text-shadow: 0 0 40px rgb(255 45 149 / 0.4);
    margin: 0.2rem 0 0.6rem;
  }
</style>
