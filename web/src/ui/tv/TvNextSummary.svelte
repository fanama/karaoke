<script lang="ts">
  import { tr } from '../../application/i18n';
  import type { QueueItem } from '../../domain/models';
  import { formatDuration } from '../../domain/playback';

  export let next: QueueItem | undefined;
</script>

<div class="next-summary">
  {#if next}
    <p class="label">{$tr('room.upNext')}</p>
    <div class="next-row">
      <img src={next.thumbnailUrl} alt="" />
      <div class="grow">
        <p class="clamp-2 next-title">{next.title}</p>
        <p class="tiny dim">{next.nickname} · {formatDuration(next.durationSeconds)}</p>
      </div>
      <span class="badge">{next.score} ↑</span>
    </div>
  {:else}
    <p class="label">{$tr('room.upNext')}</p>
    <p class="small dim">{$tr('tv.queueEmpty')}</p>
  {/if}
</div>

<style>
  .next-summary {
    text-align: right;
    min-width: 0;
  }

  .next-row {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    margin-top: 0.35rem;
  }

  .next-row img {
    width: 72px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 7px;
    order: 3;
  }

  .next-title {
    font-size: 0.82rem;
    font-weight: 700;
    line-height: 1.3;
  }

  @media (max-width: 900px) {
    .next-summary {
      display: none;
    }
  }
</style>
