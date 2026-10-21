<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { Snapshot } from '../../domain/models';
  import { computeProgress, formatDuration, voteProgress } from '../../domain/playback';

  export let snap: Snapshot;
  export let isHost: boolean;
  export let voteLeft: number;
  export let cooldownLeft: number;
  export let elapsedNow: number;
  export let onStartSkip: () => void;
  export let onCastSkip: () => void;
  export let onCancelSkip: () => void;
  export let onSearch: () => void;
</script>

{#if snap.nowPlaying}
  {@const np = snap.nowPlaying}
  <article class="now-card card">
    <div class="now-bg" style={`background-image:url(${np.thumbnailUrl})`}></div>
    <div class="now-content">
      <p class="label">{$tr('room.nowPlaying')}</p>
      <h2 class="clamp-2">{np.title}</h2>
      <p class="small muted">
        {$tr('tv.nowSinging')}: <strong>{np.nickname}</strong>
        {#if np.mine}<span class="badge badge-hot">{$tr('room.youAre')}</span>{/if}
      </p>
      <div class="progress" aria-label={$tr('room.progress')}>
        <span style={`width:${computeProgress(elapsedNow, np.durationSeconds)}%`}></span>
      </div>
      <div class="spread tiny dim">
        <span>{formatDuration(elapsedNow)}</span>
        <span>{formatDuration(np.durationSeconds)}</span>
      </div>

      <div class="skip-zone">
        {#if snap.skipVote?.active}
          <div class="vote-live">
            <div class="spread">
              <strong>{$tr('tv.skipVoteTitle')}</strong>
              <span class="mono">{voteLeft}s</span>
            </div>
            <p class="small">
              {$tr('room.needed', { yes: snap.skipVote.yes, needed: snap.skipVote.threshold })}
            </p>
            <div class="vote-bar">
              <span style={`width:${voteProgress(snap.skipVote.yes, snap.skipVote.threshold)}%`}></span>
            </div>
            {#if snap.skipVote.canVote}
              <button class="btn btn-primary btn-sm" type="button" on:click={onCastSkip}>{$tr('room.voteYes')}</button>
            {:else if snap.skipVote.votedByMe}
              <span class="badge badge-hot">{$tr('room.votedYes')}</span>
            {:else}
              <p class="tiny dim">{$tr('room.votersOnly')}</p>
            {/if}
            {#if isHost}
              <button class="btn btn-ghost btn-sm" type="button" on:click={onCancelSkip}>
                {$tr('tv.cancelVote')}
              </button>
            {/if}
          </div>
        {:else if cooldownLeft > 0}
          <button class="btn btn-secondary btn-sm" type="button" disabled>
            {$tr('room.cooldown', { seconds: cooldownLeft })}
          </button>
        {:else if !snap.you.canSkipVote}
          <button class="btn btn-secondary btn-sm" type="button" disabled>
            {$tr('room.voteToSkip')}
          </button>
          <p class="tiny dim">{$tr('room.votersOnly')}</p>
        {:else}
          <button class="btn btn-secondary btn-sm" type="button" on:click={onStartSkip}>
            {$tr('room.voteToSkip')}
          </button>
        {/if}
      </div>
    </div>
  </article>
{:else}
  <div class="empty-queue card">
    <span class="big-note">♪</span>
    <h2>{$tr('room.emptyTitle')}</h2>
    <p class="muted">{$tr('room.emptyBody')}</p>
    <button class="btn btn-primary" type="button" on:click={onSearch}>
      {$tr('room.tabSearch')}
    </button>
  </div>
{/if}

<style>
  .now-card {
    position: relative;
    overflow: hidden;
    padding: 1rem;
  }

  .now-bg {
    position: absolute;
    inset: 0;
    background-size: cover;
    background-position: center;
    opacity: 0.28;
    filter: blur(2px) saturate(1.2);
    transform: scale(1.05);
  }

  .now-card::after {
    content: "";
    position: absolute;
    inset: 0;
    background: linear-gradient(180deg, rgb(18 10 31 / 0.4), rgb(18 10 31 / 0.94));
  }

  .now-content {
    position: relative;
    z-index: 1;
    display: grid;
    gap: 0.55rem;
  }

  .now-content h2 {
    font-size: 1.25rem;
    line-height: 1.25;
  }

  .progress {
    height: 5px;
    border-radius: 999px;
    background: rgb(255 255 255 / 0.12);
    overflow: hidden;
    margin-top: 0.3rem;
  }

  .progress span {
    display: block;
    height: 100%;
    border-radius: 999px;
    background: linear-gradient(90deg, var(--magenta), var(--cyan));
  }

  .skip-zone {
    margin-top: 0.4rem;
  }

  .vote-live {
    display: grid;
    gap: 0.4rem;
    padding: 0.7rem;
    border-radius: 12px;
    background: rgb(255 45 149 / 0.08);
    border: 1px solid rgb(255 45 149 / 0.22);
  }

  .vote-bar {
    height: 6px;
    border-radius: 999px;
    background: rgb(255 255 255 / 0.12);
    overflow: hidden;
  }

  .vote-bar span {
    display: block;
    height: 100%;
    background: var(--magenta);
  }

  .empty-queue {
    display: grid;
    justify-items: center;
    gap: 0.65rem;
    text-align: center;
    padding: 2.2rem 1rem;
    color: var(--muted);
  }

  .big-note {
    font-size: 2.2rem;
    color: var(--magenta);
  }
</style>
