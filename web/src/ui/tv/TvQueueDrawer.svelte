<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { QueueItem } from '../../domain/models';

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
    <button class="tool-btn" type="button" on:click={onClose} aria-label={$tr('tv.closeQueue')}>✕</button>
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
        <span class="drag-handle" aria-hidden="true">⠿</span>
        <span class="drawer-pos mono">{index + 1}</span>
        <img src={item.thumbnailUrl} alt="" />
        <div class="grow">
          <p class="clamp-2 drawer-title">{item.title}</p>
          <p class="tiny dim">{item.nickname} · {item.score} ↑</p>
        </div>
        <button class="remove-btn" type="button" on:click={() => onRemove(item.id)} aria-label={$tr('tv.removeSong')}>✕</button>
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
    width: min(400px, 92vw);
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
    padding: 1.1rem;
    background: rgb(13 7 23 / 0.98);
    border-left: 1px solid var(--border-bright);
    box-shadow: -20px 0 80px rgb(0 0 0 / 0.5);
    transform: translateX(102%);
    transition: transform 0.23s ease;
  }

  .queue-drawer.open {
    transform: translateX(0);
  }

  .drawer-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.8rem;
  }

  .drawer-head h2 {
    font-size: 1.2rem;
  }

  .drawer-actions {
    display: flex;
    gap: 0.4rem;
  }

  .drag-hint {
    margin: -0.3rem 0;
  }

  .drawer-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.45rem;
    overflow-y: auto;
    flex: 1;
    align-content: start;
  }

  .drawer-item {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.45rem;
    border-radius: 10px;
    border: 1px solid var(--border);
    background: var(--surface);
    cursor: grab;
  }

  .drawer-item.dragging {
    opacity: 0.45;
  }

  .drag-handle {
    color: var(--dim);
    font-size: 0.9rem;
    cursor: grab;
  }

  .drawer-pos {
    color: var(--dim);
    font-size: 0.65rem;
    width: 14px;
  }

  .drawer-item img {
    width: 58px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 6px;
  }

  .drawer-title {
    font-size: 0.76rem;
    font-weight: 700;
    line-height: 1.3;
  }

  .remove-btn {
    border: 0;
    background: transparent;
    color: var(--danger);
    font-size: 0.8rem;
    width: 26px;
    height: 26px;
    border-radius: 6px;
  }

  .remove-btn:hover {
    background: rgb(255 92 92 / 0.12);
  }

  .drawer-empty {
    color: var(--dim);
    font-size: 0.8rem;
    padding: 1rem;
    text-align: center;
  }

  .drawer-footer {
    border-top: 1px solid var(--border);
    padding-top: 0.8rem;
    display: grid;
    gap: 0.5rem;
  }

  .confirm-row {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    flex-wrap: wrap;
  }

  .confirm-row .tiny {
    flex: 1;
    min-width: 120px;
  }

  .end-btn {
    justify-self: start;
  }

</style>
