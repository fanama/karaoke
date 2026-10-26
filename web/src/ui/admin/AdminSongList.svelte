<script lang="ts">
  import type { Writable } from 'svelte/store';
  import { tr } from '../../application/i18n';
  import { formatDuration } from '../../domain/playback';
  import type { AdminSong } from '../../domain/models';

  export let songs: Writable<AdminSong[]>;
  export let songsTotal: Writable<number>;
  export let songsOffset: Writable<number>;
  export let songsQuery: Writable<string>;
  export let loadingSongs: Writable<boolean>;
  export let songsError: Writable<string>;
  export let onFilter: (value: string) => void;
  export let onPrevious: () => void;
  export let onNext: () => void;
  export let onReload: () => void;

  $: lastIndex = $songsOffset + $songs.length;
  $: hasNext = lastIndex < $songsTotal;
  $: range = {
    from: $songsTotal === 0 ? 0 : $songsOffset + 1,
    to: lastIndex,
    total: $songsTotal
  };
</script>

<section class="song-list" aria-labelledby="songs-title">
  <div class="song-head">
    <div>
      <h2 id="songs-title">{$tr('admin.songsTitle')}</h2>
      {#if $songsTotal > 0}
        <p class="mono-meta">{$tr('admin.songsRange', range)}</p>
      {/if}
    </div>
    <div class="field song-filter">
      <label for="songs-filter">{$tr('admin.songsFilter')}</label>
      <input
        id="songs-filter"
        class="input"
        data-testid="songs-filter"
        bind:value={$songsQuery}
        on:input={(event) => onFilter(event.currentTarget.value)}
        placeholder={$tr('admin.songsFilterPlaceholder')}
        autocomplete="off"
        spellcheck="false"
      />
    </div>
  </div>

  {#if $songsError}
    <p class="error-text" role="alert">{$songsError}</p>
    <button class="btn btn-secondary btn-sm" type="button" on:click={onReload}>
      {$tr('common.retry')}
    </button>
  {:else if $loadingSongs}
    <p class="songs-state"><span class="spinner"></span>{$tr('admin.songsLoading')}</p>
  {:else if $songs.length === 0}
    <p class="songs-state">{$tr($songsQuery.trim() ? 'admin.songsEmpty' : 'admin.songsEmptyLibrary')}</p>
  {:else}
    <ol class="songs" data-testid="songs-list">
      {#each $songs as song, index (song.videoId)}
        <li class="song">
          <span class="song-index">{$songsOffset + index + 1}</span>
          <div class="song-copy">
            <p class="song-title clamp-2">{song.title}</p>
            <p class="song-meta">
              {#if song.channel}{song.channel} · {/if}{formatDuration(song.durationSeconds)}
            </p>
          </div>
        </li>
      {/each}
    </ol>
    <div class="pager">
      <button
        class="btn btn-secondary btn-sm"
        type="button"
        data-testid="songs-prev"
        disabled={$songsOffset === 0}
        on:click={onPrevious}
      >
        {$tr('admin.songsPrev')}
      </button>
      <span class="mono-meta">{$tr('admin.songsRange', range)}</span>
      <button
        class="btn btn-secondary btn-sm"
        type="button"
        data-testid="songs-next"
        disabled={!hasNext}
        on:click={onNext}
      >
        {$tr('admin.songsNext')}
      </button>
    </div>
  {/if}
</section>

<style>
  .song-list {
    display: grid;
    gap: var(--space-4);
    margin-top: var(--space-2);
  }

  .song-head {
    display: grid;
    gap: var(--space-4);
  }

  h2 {
    font-size: var(--text-md);
    font-weight: 800;
    letter-spacing: -0.02em;
    text-transform: uppercase;
  }

  .song-filter {
    gap: var(--space-1);
  }

  .song-filter .input {
    min-height: 44px;
    font-size: var(--text-xs);
  }

  .songs-state {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--muted);
    font-size: var(--text-xs);
  }

  .songs {
    display: grid;
    max-height: 55dvh;
    margin: 0;
    padding: 0;
    overflow-y: auto;
    list-style: none;
  }

  .song {
    display: grid;
    grid-template-columns: 4ch minmax(0, 1fr);
    gap: var(--space-3);
    align-items: baseline;
    padding: var(--space-3) 0;
  }

  .song-index {
    color: var(--dim);
    font-family: var(--mono);
    font-size: var(--text-2xs);
  }

  .song-copy {
    min-width: 0;
  }

  .song-title {
    font-size: var(--text-xs);
    font-weight: 700;
    line-height: 1.35;
  }

  .song-meta {
    color: var(--dim);
    font-size: var(--text-2xs);
  }

  .pager {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-top: var(--space-2);
  }

  @media (min-width: 700px) {
    .song-head {
      grid-template-columns: minmax(0, 1fr) 240px;
      align-items: end;
    }
  }
</style>
