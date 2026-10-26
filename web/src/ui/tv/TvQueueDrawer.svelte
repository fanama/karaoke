<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { QueueItem } from '../../domain/models';
  import Icon from '../common/Icon.svelte';

  export let queue: QueueItem[];
  export let open: boolean;
  export let manualOrder: boolean;
  export let queueLocked: boolean;
  export let confirmClear: boolean;
  export let confirmEnd: boolean;
  export let dragId: string;
  export let onClose: () => void;
  export let onToggleLock: () => void;
  export let onRotatePin: () => void;
  export let onDragStart: (event: DragEvent, id: string) => void;
  export let onDrop: (event: DragEvent, id: string) => void;
  export let onRemove: (id: string) => void;
  export let onClear: () => void;
  export let onArmClear: () => void;
  export let onArmEnd: () => void;
  export let onCancelClear: () => void;
</script>

<aside class="queue-drawer" class:open aria-hidden={!open}>
  <div class="drawer-head">
    <div>
      <h2>{$tr('tv.queueDrawer')}</h2>
      <p class="tiny dim">
        {manualOrder ? $tr('tv.orderLocked') : $tr('tv.autoOrder')}
      </p>
    </div>
    <button class="tool-btn" type="button" on:click={onClose} aria-label={$tr('tv.closeQueue')}>
      <Icon name="close" size={18} />
    </button>
  </div>

  <div class="drawer-actions">
    <button class="btn btn-secondary btn-sm" type="button" on:click={onToggleLock}>
      {$tr(queueLocked ? 'room.unlock' : 'room.lock')}
    </button>
    <button class="btn btn-ghost btn-sm" type="button" on:click={onRotatePin}>{$tr('tv.rotatePin')}</button>
  </div>
  {#if queue.length > 0}
    <p class="tiny dim drag-hint">{$tr('tv.dragHint')}</p>
  {/if}
  <ol class="drawer-list">
    {#each queue as item, index (item.id)}
      <li
        class="drawer-item"
        draggable="true"
        on:dragstart={(e) => onDragStart(e, item.id)}
        on:dragover|preventDefault
        on:drop={(e) => onDrop(e, item.id)}
        class:dragging={dragId === item.id}
      >
        <span class="drag-handle" aria-hidden="true">
          <i></i><i></i><i></i>
        </span>
        <span class="drawer-pos">{index + 1}</span>
        <img src={item.thumbnailUrl} alt="" />
        <div class="grow">
          <p class="clamp-2 drawer-title">{item.title}</p>
          <p class="drawer-meta">{item.nickname} · {item.score}</p>
        </div>
        <button
          class="remove-btn"
          type="button"
          on:click={() => onRemove(item.id)}
          aria-label={$tr('tv.removeSong')}
        >
          <Icon name="close" size={16} />
        </button>
      </li>
    {:else}
      <li class="drawer-empty">{$tr('tv.queueEmpty')}</li>
    {/each}
  </ol>

  <div class="drawer-footer">
    {#if !confirmClear}
      <button class="btn btn-ghost btn-sm" type="button" on:click={onArmClear} disabled={queue.length === 0}>
        {$tr('tv.clearQueue')}
      </button>
    {:else}
      <div class="confirm-row">
        <span class="tiny">{$tr('tv.clearConfirm')}</span>
        <button class="btn btn-danger btn-sm" type="button" on:click={onClear}>{$tr('common.confirm')}</button>
        <button class="btn btn-ghost btn-sm" type="button" on:click={onCancelClear}>{$tr('common.cancel')}</button>
      </div>
    {/if}
    <button class="btn btn-danger btn-sm end-btn" type="button" on:click={onArmEnd}>
      {confirmEnd ? $tr('tv.endRoomConfirm') : $tr('tv.endRoom')}
    </button>
  </div>
</aside>

<style>
  .queue-drawer {
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    z-index: 55;
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    width: min(400px, 92vw);
    padding: var(--space-5);
    background: var(--surface);
    box-shadow: var(--shadow-hard);
    transform: translateX(102%);
    transition: transform 0.2s ease;
  }

  .queue-drawer.open {
    transform: translateX(0);
  }

  .drawer-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .drawer-head h2 {
    font-size: var(--text-md);
    font-weight: 800;
    text-transform: uppercase;
  }

  .tool-btn {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    padding: 0;
    border: 0;
    border-radius: var(--radius-sm);
    background: var(--surface-2);
    color: var(--muted);
  }

  .tool-btn:hover {
    background: var(--surface-3);
    color: var(--text);
  }

  .drawer-actions {
    display: flex;
    gap: var(--space-1);
  }

  .drag-hint {
    margin: calc(-1 * var(--space-1)) 0;
  }

  .drawer-list {
    display: grid;
    flex: 1;
    align-content: start;
    gap: 0;
    margin: 0;
    padding: 0;
    overflow-y: auto;
    list-style: none;
  }

  .drawer-item {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) 0;
    cursor: grab;
  }

  .drawer-item.dragging {
    opacity: 0.45;
  }

  .drag-handle {
    display: grid;
    gap: var(--space-1);
    flex: none;
    cursor: grab;
  }

  .drag-handle i {
    display: block;
    width: 3px;
    height: 3px;
    background: var(--dim);
  }

  .drawer-pos {
    width: 18px;
    flex: none;
    color: var(--dim);
    font-family: var(--mono);
    font-size: var(--text-2xs);
  }

  .drawer-item img {
    width: 58px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    flex: none;
  }

  .drawer-title {
    font-size: var(--text-xs);
    font-weight: 700;
    line-height: 1.3;
  }

  .drawer-meta {
    color: var(--dim);
    font-size: var(--text-2xs);
  }

  .remove-btn {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    flex: none;
    padding: 0;
    border: 0;
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--danger);
  }

  .remove-btn:hover {
    background: var(--danger-soft);
  }

  .drawer-empty {
    padding: var(--space-4);
    color: var(--dim);
    font-size: var(--text-xs);
    text-align: center;
  }

  .drawer-footer {
    display: grid;
    gap: var(--space-3);
    margin-top: var(--space-2);
  }

  .confirm-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-1);
  }

  .confirm-row .tiny {
    flex: 1;
    min-width: 120px;
  }

  .end-btn {
    justify-self: start;
  }

</style>
