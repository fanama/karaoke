<script lang="ts">
  import type { Writable } from 'svelte/store';
  import { tr } from '../../application/i18n';
  import type { QueueSelection } from '../../domain/models';
  import Modal from '../common/Modal.svelte';

  export let selected: Writable<QueueSelection | null>;
  export let nickname: Writable<string>;
  export let nicknameError: Writable<string>;
  export let adding: Writable<boolean>;
  export let addError: Writable<string>;
  export let pinSheet: Writable<boolean>;
  export let pin: Writable<string>;
  export let pinError: Writable<string>;
  export let unlocking: Writable<boolean>;
  export let manageOpen: Writable<boolean>;
  export let confirmClear: Writable<boolean>;
  export let confirmEnd: Writable<boolean>;
  export let libraryCount: Writable<number | null>;
  export let libraryBusy: Writable<boolean>;
  export let libraryError: Writable<string>;
  export let code: string;
  export let locked: boolean;
  export let onSubmit: () => void;
  export let onUnlock: () => void;
  export let onToggleLock: () => void;
  export let onClear: () => void;
  export let onEnd: () => void;
  export let onImportLibrary: (file: File) => void;

  function pickLibraryFile(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (file) onImportLibrary(file);
    input.value = '';
  }

  function focusOnMount(node: HTMLInputElement) {
    const timer = setTimeout(() => node.focus(), 60);
    return {
      destroy() {
        clearTimeout(timer);
      }
    };
  }
</script>

<Modal open={!!$selected} title={$tr('room.nicknameTitle')} on:close={() => selected.set(null)}>
  <div class="add-sheet">
    {#if $selected?.title}
      <div class="selected-song">
        {#if $selected.thumbnailUrl}<img src={$selected.thumbnailUrl} alt="" />{/if}
        <div>
          <p class="clamp-2">{$selected.title}</p>
          {#if $selected.channel}<p class="tiny dim">{$selected.channel}</p>{/if}
        </div>
      </div>
    {/if}
    <form on:submit|preventDefault={onSubmit} class="stack">
      <div class="field">
        <label for="nickname">{$tr('room.nicknameLabel')}</label>
        <input
          id="nickname"
          class="input"
          bind:value={$nickname}
          placeholder={$tr('room.nicknamePlaceholder')}
          maxlength="24"
          autocomplete="off"
          use:focusOnMount
        />
        {#if $nicknameError}<p class="error-text">{$nicknameError}</p>{/if}
      </div>
      {#if $addError}<p class="error-text" role="alert">{$addError}</p>{/if}
      <div class="row end">
        <button class="btn btn-ghost" type="button" on:click={() => selected.set(null)}>{$tr('common.cancel')}</button>
        <button class="btn btn-primary" type="submit" disabled={$adding}>
          {#if $adding}<span class="spinner"></span>{$tr('room.adding')}{:else}{$tr('room.submit')}{/if}
        </button>
      </div>
    </form>
  </div>
</Modal>

<Modal open={$pinSheet} title={$tr('room.hostUnlockTitle')} on:close={() => pinSheet.set(false)}>
  <form class="stack" on:submit|preventDefault={onUnlock}>
    <p class="muted">{$tr('room.hostUnlockBody')}</p>
    <div class="field">
      <label for="pin">{$tr('room.pin')}</label>
      <input
        id="pin"
        class="input input-mono pin-input"
        bind:value={$pin}
        placeholder={$tr('room.pinPlaceholder')}
        maxlength="4"
        inputmode="numeric"
        pattern="[0-9]*"
        autocomplete="one-time-code"
        use:focusOnMount
      />
    </div>
    {#if $pinError}<p class="error-text" role="alert">{$pinError}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={$unlocking || $pin.trim().length !== 4}>
      {#if $unlocking}<span class="spinner"></span>{/if}{$tr('room.unlockButton')}
    </button>
  </form>
</Modal>

<Modal open={$manageOpen} title={$tr('room.manage')} on:close={() => manageOpen.set(false)}>
  <div class="stack">
    <button class="btn btn-secondary" type="button" on:click={onToggleLock}>
      {$tr(locked ? 'room.unlock' : 'room.lock')}
    </button>
    {#if !$confirmClear}
      <button class="btn btn-secondary" type="button" on:click={() => confirmClear.set(true)}>{$tr('room.clear')}</button>
    {:else}
      <div class="confirm-box">
        <p class="small">{$tr('room.clearConfirm')}</p>
        <div class="row end">
          <button class="btn btn-ghost btn-sm" type="button" on:click={() => confirmClear.set(false)}>{$tr('common.cancel')}</button>
          <button class="btn btn-danger btn-sm" type="button" on:click={onClear}>{$tr('common.confirm')}</button>
        </div>
      </div>
    {/if}
    <hr />
    <div class="library">
      <p class="label">{$tr('room.library')}</p>
      <p class="tiny dim">{$tr('room.libraryBody')}</p>
      {#if $libraryCount !== null}
        <p class="tiny dim">{$tr('room.libraryCount', { count: $libraryCount })}</p>
      {/if}
      <div class="row">
        <a class="btn btn-secondary btn-sm" href={`/api/rooms/${code}/library/export`} download>
          {$tr('room.libraryExport')}
        </a>
        <label class="btn btn-secondary btn-sm" class:disabled={$libraryBusy}>
          {$tr($libraryBusy ? 'room.libraryImporting' : 'room.libraryImport')}
          <input
            class="hidden-file"
            type="file"
            accept=".sqlite,.sqlite3,.db,application/vnd.sqlite3,application/octet-stream"
            disabled={$libraryBusy}
            on:change={pickLibraryFile}
          />
        </label>
      </div>
      {#if $libraryError}<p class="error-text">{$libraryError}</p>{/if}
    </div>
    <hr />
    {#if !$confirmEnd}
      <button class="btn btn-danger" type="button" on:click={() => confirmEnd.set(true)}>{$tr('room.endRoom')}</button>
    {:else}
      <div class="confirm-box danger">
        <p class="small">{$tr('room.endConfirm')}</p>
        <div class="row end">
          <button class="btn btn-ghost btn-sm" type="button" on:click={() => confirmEnd.set(false)}>{$tr('common.cancel')}</button>
          <button class="btn btn-danger btn-sm" type="button" on:click={onEnd}>{$tr('common.confirm')}</button>
        </div>
      </div>
    {/if}
  </div>
</Modal>

<style>
  .pin-input {
    height: 62px;
    font-family: var(--mono);
    font-size: var(--text-lg);
    letter-spacing: 0.5em;
    text-align: center;
  }

  .add-sheet {
    display: grid;
    gap: var(--space-4);
  }

  .selected-song {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2);
    border: 1px solid var(--border);
    border-radius: var(--radius-0);
    background: var(--surface-2);
  }

  .selected-song img {
    width: 82px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
  }

  .selected-song p {
    font-size: var(--text-xs);
    font-weight: 700;
    line-height: 1.3;
  }

  .row.end {
    justify-content: flex-end;
  }

  .confirm-box {
    display: grid;
    gap: var(--space-2);
    padding: var(--space-3);
    border: 1px solid var(--border);
    border-left: 2px solid var(--accent);
    border-radius: var(--radius-0);
  }

  .confirm-box.danger {
    border-left-color: var(--danger);
  }

  hr {
    border: 0;
    border-top: 1px solid var(--border);
    margin: var(--space-1) 0;
  }

  .library {
    display: grid;
    gap: var(--space-2);
  }

  .hidden-file {
    display: none;
  }
</style>
