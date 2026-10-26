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
    z-index: 20;
    display: flex;
    align-items: center;
    gap: var(--space-4);
    width: min(720px, calc(100% - 3rem));
    padding: var(--space-4) var(--space-5);
    transform: translateX(-50%);
    border-radius: var(--radius-0);
    background: var(--surface);
  }

  .skip-copy {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    font-size: var(--text-xs);
  }

  .skip-copy strong {
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .skip-copy span {
    color: var(--muted);
    font-size: var(--text-2xs);
  }

  .skip-bar {
    flex: 1;
    height: 6px;
    background: var(--surface-3);
    overflow: hidden;
  }

  .skip-bar span {
    display: block;
    height: 100%;
    background: var(--accent);
    transition: width 0.2s ease;
  }

  .skip-time {
    color: var(--accent);
    font-family: var(--mono);
    font-size: var(--text-xs);
  }

  @media (max-width: 600px) {
    .skip-banner {
      bottom: 104px;
      gap: var(--space-2);
      width: calc(100% - 1.4rem);
      flex-wrap: wrap;
    }
  }
</style>
