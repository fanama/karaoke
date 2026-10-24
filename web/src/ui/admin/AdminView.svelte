<script lang="ts">
  import { onMount } from 'svelte';
  import { tr } from '../../application/i18n';
  import { createAdminController } from '../../application/admin_controller';
  import { getAppContext } from '../../ports/context';
  import LanguageToggle from '../common/LanguageToggle.svelte';

  const c = createAdminController(getAppContext());
  const { session, token, busy, error, notice, load, login, logout, importLibrary } = c;

  onMount(() => {
    load();
  });

  function pickFile(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (file) importLibrary(file);
    input.value = '';
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
        <p class="label">{$tr('admin.library')}</p>
        {#if $session.library}
          <p class="admin-count">{$tr('admin.count', { count: $session.songs })}</p>
          <p class="tiny dim">{$tr('admin.libraryBody')}</p>
          <div class="row">
            <a class="btn btn-secondary" href="/api/admin/library/export" download>
              {$tr('admin.export')}
            </a>
            <label class="btn btn-primary" class:disabled={$busy}>
              {$tr($busy ? 'admin.importing' : 'admin.import')}
              <input
                class="hidden-file"
                type="file"
                accept=".sqlite,.sqlite3,.db,application/vnd.sqlite3,application/octet-stream"
                disabled={$busy}
                on:change={pickFile}
              />
            </label>
          </div>
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
    width: 100%;
    max-width: 560px;
    display: grid;
    gap: 1.1rem;
  }

  .admin-library {
    display: grid;
    gap: 0.75rem;
  }

  .admin-count {
    font-size: 1.5rem;
    font-weight: 800;
    color: var(--text);
  }

  .notice {
    padding: 0.8rem;
    border-radius: 12px;
    border: 1px solid rgb(34 211 238 / 0.25);
    background: rgb(34 211 238 / 0.08);
    color: #b8f6ff;
    font-size: 0.85rem;
    line-height: 1.5;
  }

  .success-text {
    color: #7ff0b0;
    font-size: 0.85rem;
    font-weight: 600;
  }

  .hidden-file {
    display: none;
  }
</style>
