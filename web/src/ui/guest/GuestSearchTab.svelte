<script lang="ts">
  import type { Writable } from 'svelte/store';
  import { tr } from '../../application/i18n';
  import { formatDuration } from '../../domain/playback';
  import type { SearchResult } from '../../domain/models';
  import Icon from '../common/Icon.svelte';
  import GuestVignettes from './GuestVignettes.svelte';

  export let q: Writable<string>;
  export let results: Writable<SearchResult[]>;
  export let suggestions: Writable<SearchResult[]>;
  export let loadingSuggestions: Writable<boolean>;
  export let searching: Writable<boolean>;
  export let searched: Writable<boolean>;
  export let stale: Writable<boolean>;
  export let searchFailed: Writable<boolean>;
  export let pasteUrl: Writable<string>;
  export let pasteError: Writable<string>;
  export let locked: boolean;
  export let full: boolean;
  export let onSearch: () => void;
  export let onClear: () => void;
  export let onChoose: (result: SearchResult) => void;
  export let onShuffle: () => void;
  export let onPaste: () => void;
</script>

<section class="search-panel">
  <div class="search-box">
    <span class="search-icon" aria-hidden="true">
      <Icon name="search" size={20} />
    </span>
    <input
      class="input search-input"
      bind:value={$q}
      on:input={onSearch}
      placeholder={$tr('room.searchPlaceholder')}
      aria-label={$tr('room.tabSearch')}
      autocomplete="off"
      enterkeyhint="search"
    />
    {#if $q}
      <button class="clear-btn" type="button" on:click={onClear} aria-label={$tr('common.close')}>
        <Icon name="close" size={16} />
      </button>
    {/if}
  </div>

  {#if locked}
    <div class="notice-strip">
      <Icon name="lock" size={16} />
      {$tr('room.queueLocked')}
    </div>
  {/if}
  {#if $stale}<div class="notice-strip cyan">{$tr('room.stale')}</div>{/if}
  {#if $searchFailed}
    <div class="notice-strip danger">{$tr('room.searchError')}</div>
  {/if}

  {#if $searching}
    <div class="searching"><span class="spinner"></span>{$tr('room.searching')}</div>
  {:else if $searched && $results.length === 0}
    <div class="empty-search">
      <Icon name="microphone" size={28} />
      <p>{$tr('room.noResults')}</p>
    </div>
  {:else if $results.length > 0}
    <ul class="results">
      {#each $results as result (result.videoId)}
        <li class="result card">
          <img src={result.thumbnailUrl} alt="" loading="lazy" width="160" height="90" />
          <div class="result-info">
            <p class="result-title clamp-2">{result.title}</p>
            <p class="tiny dim truncate">{result.channel} · {formatDuration(result.durationSeconds)}</p>
          </div>
          <button
            class="btn btn-primary btn-sm"
            type="button"
            on:click={() => onChoose(result)}
            disabled={locked || full}
          >
            {$tr('room.add')}
          </button>
        </li>
      {/each}
    </ul>
  {:else}
    <GuestVignettes
      {suggestions}
      {loadingSuggestions}
      {locked}
      {full}
      onChoose={onChoose}
      onShuffle={onShuffle}
    />
  {/if}

  <form class="paste-card card" on:submit|preventDefault={onPaste}>
    <p class="label">{$tr('room.pasteTitle')}</p>
    <div class="row">
      <input
        class="input grow"
        bind:value={$pasteUrl}
        placeholder={$tr('room.pastePlaceholder')}
        inputmode="url"
        autocomplete="off"
        disabled={locked}
      />
      <button class="btn btn-secondary" type="submit" disabled={locked || full}>
        {$tr('room.add')}
      </button>
    </div>
    {#if $pasteError}<p class="error-text">{$pasteError}</p>{/if}
  </form>
</section>

<style>
  .search-box {
    position: relative;
    margin-bottom: var(--space-3);
  }

  .search-icon {
    position: absolute;
    left: var(--space-3);
    top: 50%;
    transform: translateY(-50%);
    display: grid;
    color: var(--dim);
    pointer-events: none;
  }

  .search-input {
    height: 56px;
    padding-left: calc(var(--space-6) + var(--space-3));
    padding-right: calc(var(--space-6) + var(--space-2));
    border-radius: var(--radius-sm);
    font-size: var(--text-sm);
  }

  .clear-btn {
    position: absolute;
    right: var(--space-1);
    top: 50%;
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    padding: 0;
    transform: translateY(-50%);
    border: 0;
    background: transparent;
    color: var(--dim);
  }

  .clear-btn:hover {
    color: var(--text);
  }

  .notice-strip {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-4);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-0);
    background: var(--surface-2);
    color: var(--text);
    font-size: var(--text-xs);
    line-height: 1.4;
  }

  .notice-strip.cyan {
    background: var(--accent-soft);
    color: var(--muted);
  }

  .notice-strip.danger {
    background: var(--danger-soft);
    color: var(--danger);
  }

  .searching {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-5) var(--space-1);
    color: var(--muted);
    font-size: var(--text-xs);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .results {
    display: grid;
    gap: var(--space-3);
    margin: 0 0 var(--space-4);
    padding: 0;
    list-style: none;
  }

  .result {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3);
    background: var(--surface);
  }

  .result img {
    width: 104px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    flex: none;
    background: var(--surface-2);
  }

  .result-info {
    flex: 1;
    min-width: 0;
  }

  .result-title {
    margin-bottom: 2px;
    font-size: var(--text-xs);
    font-weight: 700;
    line-height: 1.3;
    text-wrap: pretty;
  }

  .result .btn {
    flex: none;
  }

  .empty-search {
    display: grid;
    justify-items: center;
    gap: var(--space-3);
    padding: var(--space-6) var(--space-4);
    color: var(--muted);
    text-align: center;
  }

  .empty-search :global(svg) {
    color: var(--accent);
  }

  .empty-search p {
    max-width: 34ch;
    font-size: var(--text-xs);
    line-height: 1.5;
  }

  .paste-card {
    display: grid;
    gap: var(--space-2);
    margin-top: var(--space-5);
    padding: var(--space-3);
  }

  .paste-card .input {
    min-height: 44px;
    font-size: var(--text-xs);
  }

  @media (min-width: 680px) {
    .result img {
      width: 148px;
    }
  }
</style>
