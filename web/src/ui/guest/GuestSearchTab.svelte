<script lang="ts">
  import type { Writable } from 'svelte/store';
  import { tr } from '../../application/i18n';
  import { formatDuration } from '../../domain/playback';
  import type { SearchResult } from '../../domain/models';

  export let q: Writable<string>;
  export let results: Writable<SearchResult[]>;
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
  export let onPaste: () => void;
</script>

<section class="search-panel">
  <div class="search-box">
    <span class="search-icon" aria-hidden="true">
      <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>
      </svg>
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
      <button class="clear-btn" type="button" on:click={onClear} aria-label={$tr('common.close')}>✕</button>
    {/if}
  </div>

  {#if locked}
    <div class="notice-strip">🔒 {$tr('room.queueLocked')}</div>
  {/if}
  {#if $stale}<div class="notice-strip cyan">{$tr('room.stale')}</div>{/if}
  {#if $searchFailed}
    <div class="notice-strip danger">{$tr('room.searchError')}</div>
  {/if}

  {#if $searching}
    <div class="searching"><span class="spinner"></span>{$tr('room.searching')}</div>
  {:else if $searched && $results.length === 0}
    <div class="empty-search">
      <span>♪</span>
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
          <button class="btn btn-primary btn-sm" type="button" on:click={() => onChoose(result)} disabled={locked || full}>
            {$tr('room.add')}
          </button>
        </li>
      {/each}
    </ul>
  {:else}
    <div class="search-intro">
      <div class="mic-orb" aria-hidden="true">🎤</div>
      <p>{$tr('room.searchHint')}</p>
      <p class="tiny dim">{$tr('room.searchMin')}</p>
    </div>
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
      <button class="btn btn-secondary" type="submit" disabled={locked || full}>{$tr('room.add')}</button>
    </div>
    {#if $pasteError}<p class="error-text">{$pasteError}</p>{/if}
  </form>
</section>

<style>
  .search-box {
    position: relative;
    margin-bottom: 0.9rem;
  }

  .search-icon {
    position: absolute;
    left: 0.9rem;
    top: 50%;
    transform: translateY(-50%);
    color: var(--dim);
    display: grid;
  }

  .search-input {
    padding-left: 2.7rem;
    padding-right: 2.6rem;
    height: 52px;
    border-radius: 16px;
    font-size: 1rem;
  }

  .clear-btn {
    position: absolute;
    right: 0.65rem;
    top: 50%;
    transform: translateY(-50%);
    border: 0;
    background: var(--surface-3);
    color: var(--muted);
    border-radius: 50%;
    width: 28px;
    height: 28px;
  }

  .notice-strip {
    padding: 0.65rem 0.8rem;
    margin-bottom: 0.75rem;
    border-radius: 12px;
    background: rgb(255 45 149 / 0.1);
    border: 1px solid rgb(255 45 149 / 0.25);
    color: #ffc0dd;
    font-size: 0.82rem;
  }

  .notice-strip.cyan {
    background: rgb(34 211 238 / 0.08);
    border-color: rgb(34 211 238 / 0.25);
    color: #b8f6ff;
  }

  .notice-strip.danger {
    background: rgb(255 92 92 / 0.1);
    border-color: rgb(255 92 92 / 0.28);
    color: #ffb3b3;
  }

  .searching {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    color: var(--muted);
    padding: 1.2rem 0.2rem;
    font-size: 0.9rem;
  }

  .results {
    list-style: none;
    margin: 0 0 1rem;
    padding: 0;
    display: grid;
    gap: 0.55rem;
  }

  .result {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    padding: 0.55rem;
    border-radius: 14px;
  }

  .result img {
    width: 104px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 9px;
    flex: none;
    background: var(--surface-3);
  }

  .result-info {
    flex: 1;
    min-width: 0;
  }

  .result-title {
    font-size: 0.9rem;
    font-weight: 700;
    line-height: 1.3;
    margin-bottom: 0.2rem;
  }

  .result .btn {
    flex: none;
  }

  .empty-search {
    display: grid;
    justify-items: center;
    gap: 0.65rem;
    text-align: center;
    padding: 2.2rem 1rem;
    color: var(--muted);
  }

  .empty-search span {
    font-size: 2.2rem;
    color: var(--magenta);
  }

  .empty-search p {
    max-width: 34ch;
    line-height: 1.5;
    font-size: 0.9rem;
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

  .paste-card {
    margin-top: 1.4rem;
    padding: 1rem;
    display: grid;
    gap: 0.55rem;
  }

  .paste-card .input {
    min-height: 44px;
    font-size: 0.9rem;
  }

  @media (min-width: 680px) {
    .result img {
      width: 148px;
    }
  }
</style>
