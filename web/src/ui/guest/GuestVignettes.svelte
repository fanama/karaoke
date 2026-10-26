<script lang="ts">
  import type { Writable } from 'svelte/store';
  import { tr } from '../../application/i18n';
  import { formatDuration } from '../../domain/playback';
  import type { SearchResult } from '../../domain/models';
  import Icon from '../common/Icon.svelte';

  export let suggestions: Writable<SearchResult[]>;
  export let loadingSuggestions: Writable<boolean>;
  export let locked: boolean;
  export let full: boolean;
  export let onChoose: (result: SearchResult) => void;
  export let onShuffle: () => void;
</script>

{#if $suggestions.length > 0}
  <section class="vignettes" aria-label={$tr('room.suggestionsTitle')}>
    <header class="vignettes-head">
      <p class="vignettes-title">{$tr('room.suggestionsTitle')}</p>
      <button
        class="btn btn-ghost btn-sm"
        type="button"
        on:click={onShuffle}
        disabled={$loadingSuggestions}
      >
        <Icon name="shuffle" size={14} />
        {$tr('room.shuffle')}
      </button>
    </header>

    <ul class="vignette-grid">
      {#each $suggestions as song (song.videoId)}
        <li class="vignette card">
          <div class="vignette-media">
            <img src={song.thumbnailUrl} alt="" loading="lazy" width="320" height="180" />
            <span class="duration-badge">{formatDuration(song.durationSeconds)}</span>
          </div>
          <div class="vignette-body">
            <p class="vignette-title clamp-2">{song.title}</p>
            <p class="tiny dim truncate">{song.channel}</p>
            <button
              class="btn btn-primary btn-sm"
              type="button"
              on:click={() => onChoose(song)}
              disabled={locked || full}
            >
              {$tr('room.add')}
            </button>
          </div>
        </li>
      {/each}
    </ul>
  </section>
{:else}
  <div class="search-intro">
    <span class="mic-mark" aria-hidden="true"><Icon name="microphone" size={26} /></span>
    <p>{$tr('room.searchHint')}</p>
    <p class="tiny dim">{$tr('room.searchMin')}</p>
  </div>
{/if}

<style>
  .vignettes {
    display: grid;
    gap: var(--space-3);
    margin-bottom: var(--space-5);
  }

  .vignettes-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
  }

  .vignettes-title {
    color: var(--muted);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    font-weight: 700;
    letter-spacing: 0.16em;
    text-transform: uppercase;
  }

  .vignette-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .vignette {
    display: grid;
    grid-template-rows: auto 1fr;
    padding: 0;
    overflow: hidden;
  }

  .vignette-media {
    position: relative;
    background: var(--surface-2);
  }

  .vignette-media img {
    display: block;
    width: 100%;
    aspect-ratio: 16 / 9;
    object-fit: cover;
  }

  .duration-badge {
    position: absolute;
    right: 0;
    bottom: 0;
    padding: 2px var(--space-1);
    background: var(--bg);
    color: var(--text);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    font-weight: 700;
  }

  .vignette-body {
    display: grid;
    align-content: start;
    gap: var(--space-1);
    padding: var(--space-2);
  }

  .vignette-title {
    font-size: var(--text-2xs);
    font-weight: 700;
    line-height: 1.35;
  }

  .vignette-body .btn {
    width: 100%;
    min-height: 36px;
    margin-top: var(--space-1);
  }

  .search-intro {
    display: grid;
    justify-items: center;
    gap: var(--space-2);
    padding: var(--space-6) var(--space-4);
    color: var(--muted);
    text-align: center;
  }

  .search-intro p {
    max-width: 36ch;
    font-size: var(--text-xs);
    line-height: 1.5;
  }

  .mic-mark {
    display: grid;
    place-items: center;
    width: 64px;
    height: 64px;
    margin-bottom: var(--space-1);
    color: var(--accent);
  }

  @media (min-width: 640px) {
    .vignette-grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }

    .vignette-title {
      font-size: var(--text-xs);
    }
  }
</style>
