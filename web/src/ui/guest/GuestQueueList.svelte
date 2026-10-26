<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { QueueItem } from '../../domain/models';
  import { formatDuration } from '../../domain/playback';
  import Icon from '../common/Icon.svelte';

  export let queue: QueueItem[];
  export let isHost: boolean;
  export let onVote: (entryId: string) => void;
  export let onMove: (index: number, direction: -1 | 1) => void;
  export let onRemove: (entryId: string) => void;
</script>

<div class="queue-heading">
  <h2>{$tr('room.upNext')}</h2>
  <span class="mono-meta">{$tr('room.queueCount', { count: queue.length })}</span>
</div>

{#if queue.length === 0}
  <p class="empty-line">{$tr('room.emptyBody')}</p>
{/if}

<ol class="queue-list">
  {#each queue as item, index (item.id)}
    <li class="queue-item" class:mine={item.mine}>
      <span class="position">{String(index + 1).padStart(2, '0')}</span>
      <img src={item.thumbnailUrl} alt="" loading="lazy" width="96" height="54" />
      <div class="queue-info">
        <p class="queue-title clamp-2">{item.title}</p>
        <p class="queue-meta truncate">
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
          <Icon name="up" size={14} />
          <span>{item.score}</span>
        </button>
        {#if isHost}
          <div class="host-item-actions">
            <button
              type="button"
              class="mini-action"
              on:click={() => onMove(index, -1)}
              disabled={index === 0}
              aria-label={$tr('room.moveUp')}
            >
              <Icon name="up" size={14} />
            </button>
            <button
              type="button"
              class="mini-action"
              on:click={() => onMove(index, 1)}
              disabled={index === queue.length - 1}
              aria-label={$tr('room.moveDown')}
            >
              <Icon name="down" size={14} />
            </button>
            <button
              type="button"
              class="mini-action danger"
              on:click={() => onRemove(item.id)}
              aria-label={$tr('room.remove')}
            >
              <Icon name="close" size={14} />
            </button>
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
    gap: var(--space-4);
    margin-top: var(--space-4);
  }

  .queue-heading h2 {
    font-size: var(--text-md);
    font-weight: 800;
    letter-spacing: -0.02em;
    text-transform: uppercase;
  }

  .empty-line {
    padding: var(--space-2) 0;
    color: var(--dim);
    font-size: var(--text-xs);
  }

  .queue-list {
    display: grid;
    gap: 0;
    margin: var(--space-3) 0 0;
    padding: 0;
    list-style: none;
  }

  .queue-item {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) 0;
  }

  .queue-item.mine {
    padding-inline: var(--space-2);
    background: var(--accent-soft);
  }

  .position {
    width: 20px;
    flex: none;
    color: var(--dim);
    font-family: var(--mono);
    font-size: var(--text-2xs);
  }

  .queue-item img {
    width: 82px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    flex: none;
    background: var(--surface-2);
  }

  .queue-info {
    flex: 1;
    min-width: 0;
  }

  .queue-title {
    margin-bottom: 2px;
    font-size: var(--text-xs);
    font-weight: 700;
    line-height: 1.3;
  }

  .queue-meta {
    color: var(--dim);
    font-size: var(--text-2xs);
  }

  .you {
    color: var(--accent);
    font-weight: 700;
  }

  .queue-actions {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: var(--space-1);
    flex: none;
  }

  .vote-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    min-width: 56px;
    min-height: 44px;
    padding: 0 var(--space-3);
    border: 0;
    border-radius: var(--radius-sm);
    background: var(--surface-2);
    color: var(--muted);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    font-weight: 800;
  }

  .vote-btn.active {
    background: var(--accent-soft);
    color: var(--accent);
  }

  .host-item-actions {
    display: flex;
    gap: var(--space-1);
  }

  .mini-action {
    display: grid;
    place-items: center;
    width: 40px;
    height: 32px;
    padding: 0;
    border: 0;
    border-radius: var(--radius-sm);
    background: var(--surface-2);
    color: var(--muted);
  }

  .mini-action.danger {
    color: var(--danger);
  }

  .mini-action:disabled {
    opacity: 0.3;
  }
</style>
