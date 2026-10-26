<script lang="ts">
  import { onMount } from 'svelte';
  import { tr } from '../../application/i18n';
  import { createAdminController } from '../../application/admin_controller';
  import { getAppContext } from '../../ports/context';
  import LanguageToggle from '../common/LanguageToggle.svelte';
  import AdminSongList from './AdminSongList.svelte';

  const c = createAdminController(getAppContext());
  const {
    session, token, format, busy, error, notice,
    songs, songsTotal, songsOffset, songsQuery, loadingSongs, songsError,
    load, login, logout, importLibrary, filterSongs, previousPage, nextPage, reloadSongs
  } = c;

  onMount(() => {
    load();
  });

  function pickFile(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (file) importLibrary(file);
    input.value = '';
  }

  function backendLabel(backend: string): string {
    if (backend === 'd1') return 'admin.backendD1';
    if (backend === 'sqlite') return 'admin.backendSqlite';
    return 'admin.backendNone';
  }

  function formatTime(value: string): string {
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
  }
</script>

<svelte:head>
  <title>{$tr('admin.title')} · Fana Karaoke</title>
</svelte:head>

<main class="center-page">
  <div class="card admin-card">
    <div class="spread">
      <span class="badge">Fana Karaoke</span>
      <LanguageToggle />
    </div>
    <div>
      <h1>{$tr('admin.title')}</h1>
      <p class="muted">{$tr('admin.subtitle')}</p>
    </div>

    {#if !$session}
      {#if $error}
        <p class="error-text" role="alert">{$error}</p>
      {:else}
        <p class="muted"><span class="spinner"></span> {$tr('common.loading')}</p>
      {/if}
    {:else if !$session.enabled}
      <div class="notice">{$tr('admin.disabledBody')}</div>
    {:else if !$session.authenticated}
      <form class="stack" on:submit|preventDefault={login}>
        <div class="field">
          <label for="admin-token">{$tr('admin.tokenLabel')}</label>
          <input
            id="admin-token"
            class="input input-mono"
            type="password"
            bind:value={$token}
            placeholder={$tr('admin.tokenPlaceholder')}
            autocomplete="current-password"
          />
        </div>
        {#if $error}<p class="error-text" role="alert">{$error}</p>{/if}
        <button class="btn btn-primary" type="submit" disabled={$busy || !$token.trim()}>
          {#if $busy}<span class="spinner"></span>{/if}{$tr('admin.unlock')}
        </button>
      </form>
    {:else}
      <div class="admin-library">
        {#if $session.storage}
          <div class="storage">
            <div class="spread">
              <span class="label">{$tr('admin.storage')}</span>
              <span class="badge">{$tr(backendLabel($session.storage.backend))}</span>
            </div>
            <p class="tiny dim">
              {$tr($session.storage.connected ? 'admin.connected' : 'admin.disconnected')}
              {#if $session.storage.cacheOnly} · {$tr('admin.cacheOnly')}{/if}
            </p>
            {#if $session.storage.lastSyncAt}
              <p class="tiny dim">{$tr('admin.lastSync')}: {formatTime($session.storage.lastSyncAt)}</p>
            {/if}
            {#if $session.storage.lastError}
              <p class="notice">{$tr('admin.lastError')}: {$session.storage.lastError}</p>
            {/if}
          </div>
        {/if}

        <p class="label">{$tr('admin.library')}</p>
        {#if $session.library}
          <p class="admin-count">{$tr('admin.count', { count: $session.songs })}</p>
          <p class="tiny dim">{$tr('admin.libraryBody')}</p>
          <div class="field">
            <label for="admin-format">{$tr('admin.format')}</label>
            <select id="admin-format" class="input" bind:value={$format}>
              <option value="sqlite">{$tr('admin.formatSqlite')}</option>
              <option value="sql">{$tr('admin.formatSql')}</option>
            </select>
          </div>
          <div class="row">
            <a class="btn btn-secondary" href={`/api/admin/library/export?format=${$format}`} download>
              {$tr('admin.export')}
            </a>
            <label class="btn btn-primary" class:disabled={$busy}>
              {$tr($busy ? 'admin.importing' : 'admin.import')}
              <input
                class="hidden-file"
                type="file"
                accept=".sqlite,.sqlite3,.db,.sql,.application/vnd.sqlite3,application/octet-stream,application/sql"
                disabled={$busy}
                on:change={pickFile}
              />
            </label>
          </div>

          <AdminSongList
            {songs}
            {songsTotal}
            {songsOffset}
            {songsQuery}
            {loadingSongs}
            {songsError}
            onFilter={filterSongs}
            onPrevious={previousPage}
            onNext={nextPage}
            onReload={reloadSongs}
          />
        {:else}
          <div class="notice">{$tr('admin.libraryDisabled')}</div>
        {/if}

        {#if $error}<p class="error-text" role="alert">{$error}</p>{/if}
        {#if $notice}<p class="success-text" role="status">{$notice}</p>{/if}
        <div class="row">
          <button class="btn btn-ghost btn-sm" type="button" on:click={logout}>{$tr('admin.logout')}</button>
          <a class="btn btn-ghost btn-sm" href="/">{$tr('common.home')}</a>
        </div>
      </div>
    {/if}
  </div>
</main>

<style>
  .admin-card {
    display: grid;
    gap: var(--space-6);
    width: 100%;
    max-width: 860px;
    padding: var(--space-6) var(--space-5);
    text-align: left;
  }

  .admin-card h1 {
    margin-bottom: var(--space-1);
    font-size: var(--text-xl);
    font-weight: 800;
    text-transform: uppercase;
  }

  .admin-library {
    display: grid;
    gap: var(--space-5);
  }

  .storage {
    display: grid;
    gap: var(--space-2);
    padding: var(--space-4);
    background: var(--surface-2);
  }

  .admin-count {
    color: var(--text);
    font-family: var(--mono);
    font-size: var(--text-xl);
    font-weight: 800;
  }

  .notice {
    padding: var(--space-4);
    background: var(--surface-2);
    color: var(--muted);
    font-size: var(--text-xs);
    line-height: 1.5;
  }

  .success-text {
    color: var(--accent);
    font-size: var(--text-xs);
    font-weight: 700;
  }

  .hidden-file {
    display: none;
  }

  @media (min-width: 700px) {
    .admin-card {
      padding: var(--space-6);
    }
  }
</style>
