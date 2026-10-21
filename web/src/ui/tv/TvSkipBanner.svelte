<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { SkipVote } from '../../domain/models';
  import { voteProgress } from '../../domain/playback';

  export let vote: SkipVote;
  export let voteLeft: number;
  export let onCancel: () => void;
</script>

<div class="skip-banner">
  <div class="skip-copy">
    <strong>{$tr('tv.skipVoteTitle')}</strong>
    <span>{$tr('tv.voteProgress', { yes: vote.yes, needed: vote.threshold })}</span>
  </div>
  <div class="skip-bar">
    <span style={`width:${voteProgress(vote.yes, vote.threshold)}%`}></span>
  </div>
  <span class="skip-time mono">{voteLeft}s</span>
  <button class="btn btn-ghost btn-sm" type="button" on:click={onCancel}>{$tr('tv.cancelVote')}</button>
</div>

<style>
  .skip-banner {
    position: absolute;
    left: 50%;
    bottom: 150px;
    transform: translateX(-50%);
    z-index: 20;
    display: flex;
    align-items: center;
    gap: 1rem;
    width: min(720px, calc(100% - 3rem));
    padding: 0.6rem 0.9rem;
    border-radius: 14px;
    border: 1px solid rgb(255 45 149 / 0.45);
    background: rgb(30 8 24 / 0.92);
    backdrop-filter: blur(12px);
    box-shadow: var(--shadow-neon);
  }

  .skip-copy {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
    font-size: 0.82rem;
  }

  .skip-copy span {
    color: var(--muted);
    font-size: 0.72rem;
  }

  .skip-bar {
    flex: 1;
    height: 6px;
    border-radius: 999px;
    background: rgb(255 255 255 / 0.14);
    overflow: hidden;
  }

  .skip-bar span {
    display: block;
    height: 100%;
    background: linear-gradient(90deg, var(--magenta), var(--cyan));
    transition: width 0.25s ease;
  }

  .skip-time {
    font-size: 0.85rem;
    color: var(--magenta);
  }

  @media (max-width: 600px) {
    .skip-banner {
      bottom: 104px;
      gap: 0.5rem;
      width: calc(100% - 1.4rem);
      flex-wrap: wrap;
    }
  }
</style>
