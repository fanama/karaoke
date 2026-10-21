<script lang="ts">
  import { api, ApiError } from '../lib/api';
  import { normalizeCode } from '../lib/format';
  import { errorText, tr } from '../lib/i18n';
  import LanguageToggle from '../lib/components/LanguageToggle.svelte';

  let creating = false;
  let createError = '';
  let code = '';
  let joinError = '';
  let joining = false;

  async function createRoom() {
    if (creating) return;
    creating = true;
    createError = '';
    try {
      const room = await api.createRoom();
      location.assign(room.tvUrl);
    } catch (err) {
      createError = err instanceof ApiError ? errorText($tr, err.code) : $tr('welcome.createError');
      creating = false;
    }
  }

  async function joinRoom() {
    if (joining) return;
    const normalized = normalizeCode(code);
    if (normalized.length < 3) {
      joinError = $tr('welcome.joinError');
      return;
    }
    joining = true;
    joinError = '';
    try {
      const room = await api.resolveRoom(normalized);
      location.assign(`/r/${room.code}`);
    } catch (err) {
      joinError = err instanceof ApiError ? errorText($tr, err.code) : $tr('welcome.joinError');
      joining = false;
    }
  }
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
          <button class="btn btn-primary btn-lg" type="button" on:click={createRoom} disabled={creating}>
            {#if creating}<span class="spinner"></span>{$tr('common.creating')}{:else}{$tr('common.createRoom')}{/if}
          </button>
        </div>
        {#if createError}
          <p class="error-text" role="alert">{createError}</p>
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
              bind:value={code}
              placeholder={$tr('welcome.joinPlaceholder')}
              autocomplete="off"
              autocapitalize="characters"
              spellcheck="false"
              inputmode="text"
            />
            <button class="btn btn-secondary" type="submit" disabled={joining}>
              {#if joining}<span class="spinner"></span>{/if}{$tr('common.join')}
            </button>
          </div>
          {#if joinError}<p class="error-text" role="alert">{joinError}</p>{/if}
        </form>
      </div>

      <div class="stage-card" aria-hidden="true">
        <div class="stage-screen">
          <div class="screen-top">
            <span class="screen-code">BLUE-MIC</span>
            <span class="screen-live"><i></i>LIVE</span>
          </div>
          <div class="screen-title">
            <span class="disc">♪</span>
            <div>
              <strong>Bohemian Rhapsody</strong>
              <small>Karaoke version</small>
            </div>
          </div>
          <div class="wave">
            {#each [14, 30, 22, 44, 60, 36, 72, 48, 26, 54, 34, 18, 40, 64, 28, 20, 38, 50] as h}
              <span style={`height:${h}%`}></span>
            {/each}
          </div>
          <div class="screen-bottom">
            <span class="badge badge-hot">ALEX</span>
            <span class="screen-next">Next · Dancing Queen</span>
          </div>
          <div class="qr-sim">
            {#each Array(49) as _, i}
              <i class:on={(i * 7 + Math.floor(i / 7) * 3) % 5 < 2}></i>
            {/each}
          </div>
        </div>
        <div class="stage-glow"></div>
      </div>
    </section>

    <section class="features">
      <article class="card feature">
        <span class="feature-icon magenta">⌕</span>
        <h3>{$tr('welcome.feature1Title')}</h3>
        <p class="muted">{$tr('welcome.feature1Body')}</p>
      </article>
      <article class="card feature">
        <span class="feature-icon cyan">↑</span>
        <h3>{$tr('welcome.feature2Title')}</h3>
        <p class="muted">{$tr('welcome.feature2Body')}</p>
      </article>
      <article class="card feature">
        <span class="feature-icon lime">◎</span>
        <h3>{$tr('welcome.feature3Title')}</h3>
        <p class="muted">{$tr('welcome.feature3Body')}</p>
      </article>
    </section>

    <section class="how" id="how">
      <h2>{$tr('welcome.howTitle')}</h2>
      <ol>
        <li><span>1</span>{$tr('welcome.howStep1')}</li>
        <li><span>2</span>{$tr('welcome.howStep2')}</li>
        <li><span>3</span>{$tr('welcome.howStep3')}</li>
      </ol>
    </section>

    <section class="privacy card" id="privacy">
      <div>
        <h2>{$tr('welcome.privacyTitle')}</h2>
        <p class="muted">{$tr('welcome.privacyBody')}</p>
      </div>
      <p class="tiny dim">{$tr('welcome.youtubeNote')}</p>
    </section>
  </main>

  <footer>
    <span>Fana Karaoke</span>
    <span class="dim">{$tr('app.tagline')}</span>
  </footer>
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
  main,
  footer {
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

  .stage-card {
    position: relative;
    perspective: 1200px;
  }

  .stage-screen {
    position: relative;
    aspect-ratio: 16 / 10;
    padding: 1.3rem;
    border-radius: 24px;
    border: 1px solid #542e83;
    background:
      radial-gradient(circle at 80% 0%, rgb(255 45 149 / 0.22), transparent 45%),
      radial-gradient(circle at 0% 100%, rgb(34 211 238 / 0.16), transparent 45%),
      #0c0716;
    box-shadow: 0 30px 80px rgb(0 0 0 / 0.55), 0 0 50px rgb(255 45 149 / 0.12);
    transform: rotateY(-7deg) rotateX(3deg);
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    overflow: hidden;
  }

  .screen-top,
  .screen-bottom {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .screen-code {
    font-family: var(--mono);
    font-size: clamp(1.2rem, 2.4vw, 2rem);
    font-weight: 800;
    letter-spacing: 0.12em;
  }

  .screen-live {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    color: var(--lime);
    font-size: 0.65rem;
    font-weight: 800;
    letter-spacing: 0.16em;
  }

  .screen-live i {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--lime);
    box-shadow: 0 0 10px var(--lime);
  }

  .screen-title {
    display: flex;
    align-items: center;
    gap: 0.8rem;
  }

  .screen-title strong {
    display: block;
    font-size: clamp(1.05rem, 2vw, 1.5rem);
  }

  .screen-title small {
    color: var(--muted);
  }

  .disc {
    display: grid;
    place-items: center;
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--magenta), var(--violet));
    font-size: 1.4rem;
    color: #1b0010;
  }

  .wave {
    display: flex;
    align-items: center;
    gap: 4px;
    height: 76px;
    opacity: 0.9;
  }

  .wave span {
    flex: 1;
    border-radius: 4px;
    background: linear-gradient(to top, var(--magenta), var(--cyan));
    min-height: 4px;
  }

  .qr-sim {
    position: absolute;
    right: 1.2rem;
    bottom: 2.7rem;
    width: 72px;
    height: 72px;
    padding: 6px;
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 2px;
    background: white;
    border-radius: 6px;
    opacity: 0.9;
  }

  .qr-sim i {
    background: transparent;
    border-radius: 1px;
  }

  .qr-sim i.on {
    background: #08050f;
  }

  .stage-glow {
    position: absolute;
    inset: auto 10% -30px;
    height: 70px;
    background: var(--magenta);
    filter: blur(48px);
    opacity: 0.4;
    z-index: -1;
  }

  .features {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
    padding: 1rem 0 3.5rem;
  }

  .feature {
    padding: 1.3rem;
  }

  .feature-icon {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    border-radius: 13px;
    margin-bottom: 1rem;
    font-size: 1.3rem;
    font-weight: 800;
  }

  .feature-icon.magenta {
    background: rgb(255 45 149 / 0.16);
    color: var(--magenta);
  }

  .feature-icon.cyan {
    background: rgb(34 211 238 / 0.12);
    color: var(--cyan);
  }

  .feature-icon.lime {
    background: rgb(182 255 59 / 0.1);
    color: var(--lime);
  }

  .feature h3 {
    margin-bottom: 0.4rem;
    font-size: 1.05rem;
  }

  .feature p {
    font-size: 0.9rem;
    line-height: 1.55;
  }

  .how {
    padding: 2rem 0 3rem;
  }

  .how h2,
  .privacy h2 {
    margin-bottom: 1rem;
    font-size: 1.4rem;
  }

  .how ol {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .how li {
    display: flex;
    gap: 0.8rem;
    align-items: flex-start;
    color: var(--muted);
    line-height: 1.5;
    font-size: 0.92rem;
  }

  .how li span {
    flex: none;
    display: grid;
    place-items: center;
    width: 30px;
    height: 30px;
    border-radius: 50%;
    border: 1px solid var(--border-bright);
    color: var(--cyan);
    font-family: var(--mono);
    font-size: 0.8rem;
  }

  .privacy {
    padding: 1.6rem;
    display: grid;
    gap: 1rem;
    margin-bottom: 3rem;
  }

  .privacy p {
    line-height: 1.6;
    max-width: 80ch;
    margin-top: 0.5rem;
  }

  footer {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    padding: 1.5rem 0;
    border-top: 1px solid var(--border);
    color: var(--muted);
    font-size: 0.8rem;
  }

  @media (max-width: 900px) {
    .hero {
      grid-template-columns: 1fr;
      gap: 2.5rem;
      padding-top: 1.5rem;
    }

    .stage-card {
      max-width: 540px;
    }

    .features,
    .how ol {
      grid-template-columns: 1fr;
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
    footer {
      flex-direction: column;
    }
  }
</style>
