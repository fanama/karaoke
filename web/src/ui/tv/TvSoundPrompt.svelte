<script lang="ts">
  import { tr } from '../../application/i18n';

  export let onEnable: () => void;
</script>

<button class="sound-prompt" type="button" on:click={onEnable}>
  <span class="sound-prompt-icon" aria-hidden="true">
    <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M4 9v6h4l5 4V5L8 9z"/>
      <path d="m17 9 4 6M21 9l-4 6"/>
    </svg>
  </span>
  <span class="sound-prompt-copy">
    <strong>{$tr('tv.soundOff')}</strong>
    <small>{$tr('tv.soundOffBody')}</small>
  </span>
</button>

<style>
  .sound-prompt {
    position: absolute;
    z-index: 16;
    left: 50%;
    bottom: 26%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem 1.3rem;
    border-radius: 999px;
    border: 1px solid var(--border-bright);
    background: rgb(8 5 15 / 0.82);
    backdrop-filter: blur(8px);
    color: var(--text);
    cursor: pointer;
    box-shadow: 0 18px 50px rgb(0 0 0 / 0.55), var(--shadow-neon);
    animation: sound-pulse 2.4s ease-in-out infinite;
  }

  .sound-prompt-icon {
    display: grid;
    place-items: center;
    width: 42px;
    height: 42px;
    border-radius: 50%;
    border: 1px solid var(--magenta);
    color: var(--magenta);
    flex: none;
  }

  .sound-prompt-copy {
    display: grid;
    text-align: left;
    gap: 0.1rem;
  }

  .sound-prompt-copy strong {
    font-size: 1rem;
  }

  .sound-prompt-copy small {
    color: var(--muted);
    font-size: 0.78rem;
  }

  @keyframes sound-pulse {
    50% {
      box-shadow: 0 18px 50px rgb(0 0 0 / 0.55), 0 0 34px rgb(255 45 149 / 0.45);
    }
  }
</style>
