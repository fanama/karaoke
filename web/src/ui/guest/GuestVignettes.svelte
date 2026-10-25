<script lang="ts">
  import type { Writable } from 'svelte/store';
  import { tr } from '../../application/i18n';
  import { formatDuration } from '../../domain/playback';
  import type { SearchResult } from '../../domain/models';

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
        class="btn btn-ghost btn-sm shuffle"
        type="button"
        on:click={onShuffle}
        disabled={$loadingSuggestions}
      >
        <span class="shuffle-icon" aria-hidden="true">⤨</span>
        {$tr('room.shuffle')}
      </button>
    </header>

    <ul class="vignette-grid">
      {#each $suggestions as song (song.videoId)}
        <li class="vignette card">
          <div class="vignette-media">
            <img
              src={song.thumbnailUrl}
              alt=""
              loading="lazy"
              width="320"
              height="180"
            />
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
    <div class="mic-orb" aria-hidden="true">🎤</div>
    <p>{$tr('room.searchHint')}</p>
    <p class="tiny dim">{$tr('room.searchMin')}</p>
  </div>
{/if}

<style>
  .vignettes {
    display: grid;
    gap: 0.7rem;
    margin-bottom: 1.2rem;
  }

  .vignettes-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.6rem;
  }

  .vignettes-title {
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--cyan);
  }

  .shuffle {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    min-height: 36px;
    padding: 0 0.7rem;
  }

  .shuffle-icon {
    font-size: 1rem;
    line-height: 1;
  }

  .vignette-grid {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.65rem;
  }

  .vignette {
    display: grid;
    grid-template-rows: auto 1fr;
    padding: 0;
    overflow: hidden;
    border-radius: 14px;
  }

  .vignette-media {
    position: relative;
    background: var(--surface-3);
  }

  .vignette-media img {
    display: block;
    width: 100%;
    aspect-ratio: 16 / 9;
    object-fit: cover;
  }

  .duration-badge {
    position: absolute;
    right: 0.35rem;
    bottom: 0.35rem;
    padding: 0.1rem 0.35rem;
    border-radius: 6px;
    background: rgb(8 5 15 / 0.82);
    font-family: var(--mono);
    font-size: 0.7rem;
    font-weight: 700;
    color: var(--text);
  }

  .vignette-body {
    display: grid;
    gap: 0.3rem;
    padding: 0.55rem 0.6rem 0.7rem;
    align-content: start;
  }

  .vignette-title {
    font-size: 0.82rem;
    font-weight: 700;
    line-height: 1.3;
  }

  .vignette-body .btn {
    margin-top: 0.25rem;
    width: 100%;
    min-height: 38px;
  }

  .search-intro {
    display: grid;
    justify-items: center;
    gap: 0.6rem;
    text-align: center;
    padding: 2.4rem 1rem;
    color: var(--muted);
  }

  .search-intro p {
    max-width: 36ch;
    line-height: 1.5;
    font-size: 0.92rem;
  }

  .mic-orb {
    display: grid;
    place-items: center;
    width: 72px;
    height: 72px;
    border-radius: 50%;
    background: radial-gradient(circle at 30% 30%, rgb(255 45 149 / 0.35), rgb(139 92 246 / 0.12));
    border: 1px solid var(--border-bright);
    box-shadow: 0 0 40px rgb(255 45 149 / 0.18);
    font-size: 1.8rem;
    margin-bottom: 0.3rem;
  }

  @media (min-width: 640px) {
    .vignette-grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
  }
</style>
