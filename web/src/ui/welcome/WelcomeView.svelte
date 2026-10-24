<script lang="ts">
  import { tr } from '../../application/i18n';
  import { createWelcomeController } from '../../application/welcome_controller';
  import { getAppContext } from '../../ports/context';
  import LanguageToggle from '../common/LanguageToggle.svelte';
  import WelcomeFeatures from './WelcomeFeatures.svelte';
  import WelcomeFooter from './WelcomeFooter.svelte';
  import WelcomeStagePreview from './WelcomeStagePreview.svelte';

  const { creating, createError, code, joinError, joining, createRoom, joinRoom } =
    createWelcomeController(getAppContext());
</script>

<svelte:head>
  <title>Fana Karaoke — browser karaoke rooms</title>
</svelte:head>

<div class="welcome">
  <div class="glow glow-a"></div>
  <div class="glow glow-b"></div>

  <header class="topbar">
    <a class="brand" href="/" aria-label="Fana Karaoke">
      <span class="brand-mark" aria-hidden="true">
        <svg viewBox="0 0 32 32" width="28" height="28">
          <path d="M12 23V9l12-3v13" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
          <circle cx="9" cy="24" r="4" fill="var(--cyan)"/>
          <circle cx="21" cy="20" r="4" fill="var(--magenta)"/>
        </svg>
      </span>
      <span>Fana <strong>Karaoke</strong></span>
    </a>
    <div class="row">
      <a class="btn btn-ghost btn-sm desktop-only" href="#how">{$tr('welcome.howOpen')}</a>
      <a class="btn btn-ghost btn-sm desktop-only" href="#privacy">{$tr('welcome.privacyOpen')}</a>
      <LanguageToggle />
    </div>
  </header>

  <main>
    <section class="hero">
      <div class="hero-copy">
        <p class="eyebrow"><span></span>{$tr('welcome.eyebrow')}</p>
        <h1>{$tr('welcome.title')}</h1>
        <p class="lede">{$tr('welcome.subtitle')}</p>

        <div class="actions">
          <button class="btn btn-primary btn-lg" type="button" on:click={createRoom} disabled={$creating}>
            {#if $creating}<span class="spinner"></span>{$tr('common.creating')}{:else}{$tr('common.createRoom')}{/if}
          </button>
        </div>
        {#if $createError}
          <p class="error-text" role="alert">{$createError}</p>
        {/if}

        <form class="join-card card" on:submit|preventDefault={joinRoom}>
          <div>
            <h2>{$tr('welcome.joinTitle')}</h2>
            <p class="small muted">{$tr('welcome.joinHint')}</p>
          </div>
          <div class="join-row">
            <label class="visually-hidden" for="room-code">{$tr('welcome.joinLabel')}</label>
            <input
              id="room-code"
              class="input input-mono"
              bind:value={$code}
              placeholder={$tr('welcome.joinPlaceholder')}
              autocomplete="off"
              autocapitalize="characters"
              spellcheck="false"
              inputmode="text"
            />
            <button class="btn btn-secondary" type="submit" disabled={$joining}>
              {#if $joining}<span class="spinner"></span>{/if}{$tr('common.join')}
            </button>
          </div>
          {#if $joinError}<p class="error-text" role="alert">{$joinError}</p>{/if}
        </form>
      </div>

      <WelcomeStagePreview />
    </section>

    <WelcomeFeatures />
  </main>

  <WelcomeFooter />
</div>

<style>
  .welcome {
    position: relative;
    overflow: hidden;
    min-height: 100dvh;
    padding: 0 1.25rem 2rem;
  }

  .glow {
    position: absolute;
    border-radius: 50%;
    filter: blur(80px);
    pointer-events: none;
    opacity: 0.45;
    z-index: 0;
  }

  .glow-a {
    width: 520px;
    height: 520px;
    background: #9d1d64;
    top: -220px;
    right: -120px;
  }

  .glow-b {
    width: 460px;
    height: 460px;
    background: #0e5d70;
    bottom: -260px;
    left: -160px;
    opacity: 0.3;
  }

  .topbar,
  main {
    position: relative;
    z-index: 1;
    max-width: 1180px;
    margin: 0 auto;
  }

  .topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 1.4rem 0;
  }

  .brand {
    display: inline-flex;
    align-items: center;
    gap: 0.6rem;
    color: var(--text);
    text-decoration: none;
    font-size: 1.05rem;
    letter-spacing: -0.02em;
  }

  .brand strong {
    color: var(--magenta);
  }

  .brand-mark {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    border-radius: 12px;
    border: 1px solid var(--border-bright);
    background: var(--surface);
    color: var(--magenta);
  }

  .hero {
    display: grid;
    grid-template-columns: 1.05fr 0.95fr;
    gap: 3rem;
    align-items: center;
    min-height: 66vh;
    padding: 3rem 0 4rem;
  }

  .eyebrow {
    display: inline-flex;
    align-items: center;
    gap: 0.55rem;
    margin-bottom: 1rem;
    color: var(--cyan);
    font-size: 0.75rem;
    font-weight: 800;
    letter-spacing: 0.18em;
    text-transform: uppercase;
  }

  .eyebrow span {
    width: 26px;
    height: 2px;
    background: var(--cyan);
  }

  h1 {
    font-size: clamp(2.6rem, 5.4vw, 4.6rem);
    line-height: 0.98;
    font-weight: 900;
    max-width: 13ch;
    background: linear-gradient(120deg, #fff 35%, #ffc2df 70%, var(--magenta));
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
  }

  .lede {
    margin: 1.5rem 0 1.8rem;
    max-width: 46ch;
    color: var(--muted);
    font-size: 1.08rem;
    line-height: 1.6;
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.8rem;
    margin-bottom: 0.7rem;
  }

  .join-card {
    margin-top: 2rem;
    padding: 1.1rem;
    display: grid;
    gap: 0.85rem;
    max-width: 480px;
  }

  .join-card h2 {
    font-size: 1rem;
  }

  .join-row {
    display: flex;
    gap: 0.5rem;
  }

  .join-row .input {
    flex: 1;
  }

  @media (max-width: 900px) {
    .hero {
      grid-template-columns: 1fr;
      gap: 2.5rem;
      padding-top: 1.5rem;
    }

    .desktop-only {
      display: none;
    }
  }

  @media (max-width: 420px) {
    .join-row {
      flex-direction: column;
    }
    .join-row .btn {
      width: 100%;
    }
  }
</style>
