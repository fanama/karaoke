<script lang="ts">
  import { tr } from '../../application/i18n';
  import { createWelcomeController } from '../../application/welcome_controller';
  import { getAppContext } from '../../ports/context';
  import Icon from '../common/Icon.svelte';
  import LanguageToggle from '../common/LanguageToggle.svelte';
  import WelcomeFooter from './WelcomeFooter.svelte';
  import WelcomeJoinTicket from './WelcomeJoinTicket.svelte';
  import WelcomeLineup from './WelcomeLineup.svelte';
  import WelcomePrivacy from './WelcomePrivacy.svelte';
  import WelcomeProgramme from './WelcomeProgramme.svelte';
  import WelcomeStagePreview from './WelcomeStagePreview.svelte';

  const { creating, createError, code, joinError, joining, createRoom, joinRoom } =
    createWelcomeController(getAppContext());
</script>

<svelte:head>
  <title>Fana Karaoke — browser karaoke rooms</title>
</svelte:head>

<div class="welcome">
  <header class="masthead">
    <a class="brand" href="/" aria-label="Fana Karaoke">
      <span class="brand-mark" aria-hidden="true">
        <svg viewBox="0 0 32 32" width="26" height="26">
          <path
            d="M12 23V9l12-3v13"
            fill="none"
            stroke="currentColor"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <circle cx="9" cy="24" r="4" fill="var(--accent)" />
          <circle cx="21" cy="20" r="4" fill="var(--accent)" />
        </svg>
      </span>
      <span class="brand-name">Fana <strong>Karaoke</strong></span>
    </a>
    <nav class="masthead-nav" aria-label="About">
      <a href="#how">{$tr('welcome.howOpen')}</a>
      <a href="#privacy">{$tr('welcome.privacyOpen')}</a>
    </nav>
    <LanguageToggle />
  </header>

  <main>
    <section class="hero">
      <div class="hero-copy">
        <p class="eyebrow">{$tr('welcome.eyebrow')}</p>
        <h1>{$tr('welcome.title')}</h1>
        <p class="lede">{$tr('welcome.subtitle')}</p>

        <div class="actions">
          <button
            class="btn btn-primary btn-lg"
            type="button"
            data-testid="create-room"
            on:click={createRoom}
            disabled={$creating}
          >
            {#if $creating}<span class="spinner"></span>{$tr('common.creating')}{:else}
              {$tr('common.createRoom')}
            {/if}
          </button>
          <a class="btn btn-secondary btn-lg" href="#how">
            {$tr('welcome.howOpen')}
            <Icon name="chevron" size={14} />
          </a>
        </div>
        {#if $createError}
          <p class="error-text" role="alert">{$createError}</p>
        {/if}
      </div>

      <WelcomeJoinTicket {code} {joining} {joinError} onJoin={joinRoom} />
    </section>

    <WelcomeStagePreview />
    <WelcomeLineup />
    <WelcomeProgramme />
    <WelcomePrivacy />
  </main>

  <WelcomeFooter />
</div>

<style>
  .welcome {
    min-height: 100dvh;
    max-width: 1180px;
    margin: 0 auto;
    padding: 0 var(--space-5) var(--space-6);
  }

  .masthead {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-3) 0;
    border-bottom: 1px solid var(--border);
  }

  .brand {
    display: inline-flex;
    align-items: center;
    gap: var(--space-3);
    color: var(--text);
    text-decoration: none;
  }

  .brand-mark {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    color: var(--text);
  }

  .brand-name {
    font-size: var(--text-sm);
    font-weight: 800;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .brand-name strong {
    color: var(--accent);
  }

  .masthead-nav {
    display: none;
    align-items: center;
    gap: var(--space-5);
    margin-left: auto;
  }

  .masthead-nav a {
    display: inline-flex;
    align-items: center;
    min-height: 44px;
    color: var(--muted);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    letter-spacing: 0.1em;
    text-transform: uppercase;
    text-decoration: none;
    border-bottom: 2px solid transparent;
  }

  .masthead-nav a:hover {
    color: var(--text);
    border-bottom-color: var(--accent);
  }

  .masthead :global(.lang-toggle) {
    margin-left: auto;
  }

  .hero {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    border: 1px solid var(--border);
    margin-top: var(--space-5);
  }

  .hero-copy {
    padding: var(--space-5);
  }

  .hero h1 {
    margin: var(--space-4) 0 0;
    max-width: 11ch;
    font-size: clamp(2.75rem, 6.2vw, 5.5rem);
    font-weight: 800;
    line-height: 0.95;
    letter-spacing: -0.035em;
    text-transform: uppercase;
    text-wrap: balance;
  }

  .lede {
    margin-top: var(--space-5);
    max-width: 46ch;
    color: var(--muted);
    font-size: var(--text-sm);
    line-height: 1.55;
  }

  .actions {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-top: var(--space-5);
  }

  .actions .btn {
    width: 100%;
  }

  main :global(section) {
    margin-top: var(--space-6);
  }

  @media (min-width: 900px) {
    .welcome {
      padding-inline: var(--space-6);
    }

    .masthead {
      padding-block: var(--space-4);
    }

    .masthead-nav {
      display: flex;
    }

    .masthead :global(.lang-toggle) {
      margin-left: 0;
    }

    .hero {
      grid-template-columns: minmax(0, 1fr) 420px;
      margin-top: var(--space-6);
    }

    .hero-copy {
      padding: var(--space-7) var(--space-6);
    }

    .actions {
      flex-direction: row;
      align-items: center;
    }

    .actions .btn {
      width: auto;
    }

    main :global(section) {
      margin-top: var(--space-7);
    }
  }
</style>
