<script lang="ts">
  import { tr } from '../../application/i18n';
  import Icon from '../common/Icon.svelte';

  export let onEnable: () => void;
</script>

<button class="sound-prompt" type="button" on:click={onEnable}>
  <span class="sound-prompt-icon" aria-hidden="true"><Icon name="volume" size={22} /></span>
  <span class="sound-prompt-copy">
    <strong>{$tr('tv.soundOff')}</strong>
    <small>{$tr('tv.soundOffBody')}</small>
  </span>
</button>

<style>
  .sound-prompt {
    position: absolute;
    left: 50%;
    bottom: 26%;
    z-index: 16;
    display: flex;
    align-items: center;
    gap: var(--space-3);
    max-width: min(90vw, 460px);
    padding: var(--space-3) var(--space-4);
    transform: translateX(-50%);
    border: 1px solid var(--border-strong);
    border-left: 2px solid var(--accent);
    border-radius: var(--radius-0);
    background: var(--surface);
    color: var(--text);
    text-align: left;
  }

  .sound-prompt:hover {
    border-color: var(--accent);
  }

  .sound-prompt-icon {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    flex: none;
    border: 1px solid var(--accent);
    color: var(--accent);
  }

  .sound-prompt-copy {
    display: grid;
    gap: var(--space-1);
  }

  .sound-prompt-copy strong {
    font-size: var(--text-sm);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .sound-prompt-copy small {
    color: var(--muted);
    font-size: var(--text-2xs);
    line-height: 1.4;
  }
</style>
