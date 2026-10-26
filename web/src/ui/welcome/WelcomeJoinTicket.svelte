<script lang="ts">
  import type { Writable } from 'svelte/store';
  import { tr } from '../../application/i18n';

  export let code: Writable<string>;
  export let joining: Writable<boolean>;
  export let joinError: Writable<string>;
  export let onJoin: () => void;
</script>

<form class="ticket paper" on:submit|preventDefault={onJoin} data-testid="join-form">
  <h2>{$tr('welcome.joinTitle')}</h2>

  <label class="label" for="room-code">{$tr('welcome.joinLabel')}</label>
  <input
    id="room-code"
    class="input input-mono ticket-input"
    data-testid="room-code-input"
    bind:value={$code}
    placeholder={$tr('welcome.joinPlaceholder')}
    autocomplete="off"
    autocapitalize="characters"
    spellcheck="false"
    inputmode="text"
  />
  <button class="btn ticket-submit" type="submit" disabled={$joining}>
    {#if $joining}<span class="spinner"></span>{/if}{$tr('common.join')}
  </button>

  {#if $joinError}
    <p class="ticket-error" role="alert">{$joinError}</p>
  {/if}
  <p class="ticket-hint">{$tr('welcome.joinHint')}</p>

  <ul class="stub">
    <li>{$tr('welcome.ticketNoAccounts')}</li>
    <li>{$tr('welcome.ticketInMemory')}</li>
  </ul>
</form>

<style>
  .ticket {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    height: 100%;
    padding: var(--space-5);
    background: var(--paper);
    border-radius: var(--radius-0);
    color: var(--ink);
  }

  .ticket h2 {
    font-size: var(--text-md);
    font-weight: 800;
    letter-spacing: -0.01em;
    text-transform: uppercase;
  }

  .label {
    margin-top: var(--space-3);
    color: var(--ink-muted);
  }

  .ticket-input {
    min-height: 56px;
    border: 1px solid var(--ink);
    border-radius: var(--radius-0);
    background: transparent;
    color: var(--ink);
    font-size: var(--text-md);
  }

  .ticket-input::placeholder {
    color: var(--ink-muted);
  }

  .ticket-input:focus {
    border-color: var(--ink);
    outline: 2px solid var(--ink);
    outline-offset: 2px;
  }

  .ticket-submit {
    min-height: 52px;
    margin-top: var(--space-2);
    border-radius: var(--radius-0);
    background: var(--ink);
    color: var(--paper);
  }

  .ticket-submit:hover:not(:disabled) {
    background: var(--ink);
    color: var(--accent);
    transform: none;
  }

  .ticket-error {
    padding: var(--space-2) var(--space-3);
    background: var(--danger-soft);
    color: var(--ink);
    font-size: var(--text-xs);
    font-weight: 700;
  }

  .ticket-hint {
    color: var(--ink-muted);
    font-size: var(--text-xs);
    line-height: 1.45;
  }

  .stub {
    display: grid;
    gap: var(--space-3);
    margin: var(--space-5) 0 0;
    padding: 0;
    list-style: none;
  }

  .stub li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--ink-muted);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    letter-spacing: 0.14em;
    text-transform: uppercase;
  }

  .stub li::before {
    content: "";
    width: 6px;
    height: 6px;
    background: var(--accent);
  }

  @media (min-width: 900px) {
    .ticket {
      padding: var(--space-6);
      gap: var(--space-4);
    }
  }
</style>
