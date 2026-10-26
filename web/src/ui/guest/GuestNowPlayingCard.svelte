<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { Snapshot } from '../../domain/models';
  import { computeProgress, formatDuration, voteProgress } from '../../domain/playback';
  import Icon from '../common/Icon.svelte';

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
  <article class="now-card">
    <div class="now-head">
      <img src={np.thumbnailUrl} alt="" width="160" height="90" />
      <div class="now-copy">
        <p class="label">{$tr('room.nowPlaying')}</p>
        <h2 class="clamp-2">{np.title}</h2>
        <p class="now-singer">
          {$tr('tv.nowSinging')}: <strong>{np.nickname}</strong>
        </p>
        {#if np.mine}<span class="badge badge-hot">{$tr('room.youAre')}</span>{/if}
      </div>
    </div>

    <div class="progress" aria-label={$tr('room.progress')}>
      <span style={`width:${computeProgress(elapsedNow, np.durationSeconds)}%`}></span>
    </div>
    <div class="spread mono-meta">
      <span>{formatDuration(elapsedNow)}</span>
      <span>{formatDuration(np.durationSeconds)}</span>
    </div>

    <div class="skip-zone">
      {#if snap.skipVote?.active}
        <div class="vote-live">
          <div class="spread">
            <strong>{$tr('tv.skipVoteTitle')}</strong>
            <span class="mono-meta">{voteLeft}s</span>
          </div>
          <p class="small muted">
            {$tr('room.needed', { yes: snap.skipVote.yes, needed: snap.skipVote.threshold })}
          </p>
          <div class="vote-bar">
            <span style={`width:${voteProgress(snap.skipVote.yes, snap.skipVote.threshold)}%`}></span>
          </div>
          {#if snap.skipVote.canVote}
            <button class="btn btn-primary btn-sm" type="button" on:click={onCastSkip}>
              {$tr('room.voteYes')}
            </button>
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
  </article>
{:else}
  <div class="empty-queue">
    <span class="empty-mark" aria-hidden="true"><Icon name="microphone" size={26} /></span>
    <h2>{$tr('room.emptyTitle')}</h2>
    <p class="muted">{$tr('room.emptyBody')}</p>
    <button class="btn btn-primary" type="button" on:click={onSearch}>
      {$tr('room.tabSearch')}
    </button>
  </div>
{/if}

<style>
  .now-card {
    display: grid;
    gap: var(--space-2);
    padding: var(--space-4);
    border: 1px solid var(--border);
    border-top: 2px solid var(--accent);
  }

  .now-head {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
  }

  .now-head img {
    width: 104px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    flex: none;
    background: var(--surface-2);
  }

  .now-copy {
    display: grid;
    gap: var(--space-1);
    min-width: 0;
    justify-items: start;
  }

  .now-copy h2 {
    font-size: var(--text-md);
    font-weight: 800;
    line-height: 1.2;
    letter-spacing: -0.02em;
    text-wrap: balance;
  }

  .now-singer {
    color: var(--muted);
    font-size: var(--text-xs);
  }

  .progress {
    height: 4px;
    margin-top: var(--space-2);
    background: var(--surface-3);
    overflow: hidden;
  }

  .progress span {
    display: block;
    height: 100%;
    background: var(--accent);
  }

  .skip-zone {
    margin-top: var(--space-2);
  }

  .vote-live {
    display: grid;
    gap: var(--space-2);
    padding: var(--space-3);
    border: 1px solid var(--border);
    border-left: 2px solid var(--accent);
  }

  .vote-live strong {
    font-size: var(--text-xs);
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .vote-bar {
    height: 4px;
    background: var(--surface-3);
    overflow: hidden;
  }

  .vote-bar span {
    display: block;
    height: 100%;
    background: var(--accent);
  }

  .empty-queue {
    display: grid;
    justify-items: center;
    gap: var(--space-2);
    padding: var(--space-6) var(--space-4);
    border: 1px solid var(--border);
    color: var(--muted);
    text-align: center;
  }

  .empty-queue h2 {
    font-size: var(--text-md);
    font-weight: 800;
    text-transform: uppercase;
  }

  .empty-mark {
    display: grid;
    place-items: center;
    width: 56px;
    height: 56px;
    margin-bottom: var(--space-1);
    border: 1px solid var(--accent);
    color: var(--accent);
  }
</style>
