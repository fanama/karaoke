<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { QueueItem } from '../../domain/models';
  import { formatDuration } from '../../domain/playback';

  export let queue: QueueItem[];
  export let isHost: boolean;
  export let onVote: (entryId: string) => void;
  export let onMove: (index: number, direction: -1 | 1) => void;
  export let onRemove: (entryId: string) => void;
</script>

<div class="queue-heading">
  <h2>{$tr('room.upNext')}</h2>
  <span class="tiny dim">{$tr('room.queueCount', { count: queue.length })}</span>
</div>

{#if queue.length === 0}
  <p class="tiny dim empty-line">{$tr('room.emptyBody')}</p>
{/if}

<ol class="queue-list">
  {#each queue as item, index (item.id)}
    <li class="queue-item card" class:mine={item.mine}>
      <span class="position mono">{index + 1}</span>
      <img src={item.thumbnailUrl} alt="" loading="lazy" width="96" height="54" />
      <div class="queue-info">
        <p class="queue-title clamp-2">{item.title}</p>
        <p class="tiny dim truncate">
          {item.nickname}{#if item.mine} · <span class="you">{$tr('room.youAre')}</span>{/if}
          · {formatDuration(item.durationSeconds)}
        </p>
      </div>
      <div class="queue-actions">
        <button
          class:active={item.votedByMe}
          class="vote-btn"
          type="button"
          on:click={() => onVote(item.id)}
          aria-label={$tr(item.votedByMe ? 'room.voted' : 'room.vote')}
          title={$tr(item.votedByMe ? 'room.voted' : 'room.vote')}
        >
          <svg viewBox="0 0 24 24" width="15" height="15" fill={item.votedByMe ? 'currentColor' : 'none'} stroke="currentColor" stroke-width="2">
            <path d="M12 19V5M5 12l7-7 7 7"/>
          </svg>
          <span>{item.score}</span>
        </button>
        {#if isHost}
          <div class="host-item-actions">
            <button type="button" class="mini-action" on:click={() => onMove(index, -1)} disabled={index === 0} aria-label={$tr('room.moveUp')}>↑</button>
            <button type="button" class="mini-action" on:click={() => onMove(index, 1)} disabled={index === queue.length - 1} aria-label={$tr('room.moveDown')}>↓</button>
            <button type="button" class="mini-action danger" on:click={() => onRemove(item.id)} aria-label={$tr('room.remove')}>✕</button>
          </div>
        {/if}
      </div>
    </li>
  {/each}
</ol>

<style>
  .queue-heading {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 1rem;
    margin-top: 0.4rem;
  }

  .queue-heading h2 {
    font-size: 1.05rem;
  }

  .empty-line {
    padding: 0.4rem 0;
  }

  .queue-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.55rem;
  }

  .queue-item {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    padding: 0.55rem;
    border-radius: 14px;
  }

  .queue-item.mine {
    border-color: rgb(255 45 149 / 0.35);
  }

  .position {
    width: 18px;
    flex: none;
    color: var(--dim);
    font-size: 0.72rem;
    text-align: center;
  }

  .queue-item img {
    width: 82px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 8px;
    flex: none;
    background: var(--surface-3);
  }

  .queue-info {
    flex: 1;
    min-width: 0;
  }

  .queue-title {
    font-size: 0.88rem;
    font-weight: 700;
    line-height: 1.3;
    margin-bottom: 0.2rem;
  }

  .you {
    color: var(--magenta);
    font-weight: 700;
  }

  .queue-actions {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.3rem;
    flex: none;
  }

  .vote-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    min-width: 46px;
    min-height: 34px;
    padding: 0 0.5rem;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--muted);
    font-size: 0.78rem;
    font-weight: 800;
  }

  .vote-btn.active {
    border-color: var(--magenta);
    background: rgb(255 45 149 / 0.14);
    color: #ffb3d7;
  }

  .host-item-actions {
    display: flex;
    gap: 2px;
  }

  .mini-action {
    width: 26px;
    height: 26px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--muted);
    font-size: 0.75rem;
    padding: 0;
  }

  .mini-action.danger {
    color: var(--danger);
  }

  .mini-action:disabled {
    opacity: 0.3;
  }
</style>
