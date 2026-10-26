<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { tr } from '../../application/i18n';
  import Icon from './Icon.svelte';

  export let open = false;
  export let title = '';
  export let dismissible = true;

  const dispatch = createEventDispatcher<{ close: void }>();
  let panel: HTMLDivElement;

  function close() {
    if (dismissible) dispatch('close');
  }

  function onKey(event: KeyboardEvent) {
    if (event.key === 'Escape' && open) close();
  }

  onMount(() => {
    const previous = document.body.style.overflow;
    return () => {
      document.body.style.overflow = previous;
    };
  });

  $: if (open) {
    document.body.style.overflow = 'hidden';
    setTimeout(() => panel?.focus(), 0);
  } else {
    document.body.style.overflow = '';
  }
</script>

<svelte:window on:keydown={onKey} />

{#if open}
  <div class="modal-backdrop" role="presentation" on:click|self={close}>
    <div class="modal" role="dialog" aria-modal="true" aria-label={title} tabindex="-1" bind:this={panel}>
      {#if title}
        <div class="spread" style="margin-bottom: 1rem; align-items: flex-start;">
          <h2>{title}</h2>
          {#if dismissible}
            <button
              class="btn btn-ghost btn-icon"
              type="button"
              on:click={close}
              aria-label={$tr('common.close')}
            >
              <Icon name="close" size={18} />
            </button>
          {/if}
        </div>
      {/if}
      <slot />
    </div>
  </div>
{/if}
