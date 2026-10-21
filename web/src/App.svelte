<script lang="ts">
  import { onMount } from 'svelte';
  import Welcome from './views/Welcome.svelte';
  import GuestRoom from './views/GuestRoom.svelte';
  import TvRoom from './views/TvRoom.svelte';
  import LanguageToggle from './lib/components/LanguageToggle.svelte';
  import { tr } from './lib/i18n';

  let path = location.pathname.replace(/\/+$/, '') || '/';

  function syncPath() {
    path = location.pathname.replace(/\/+$/, '') || '/';
  }

  onMount(() => {
    window.addEventListener('popstate', syncPath);
    return () => window.removeEventListener('popstate', syncPath);
  });

  $: guest = path.match(/^\/r\/([^/]+)$/);
  $: tv = path.match(/^\/tv\/([^/]+)$/);
</script>

{#if path === '/'}
  <Welcome />
{:else if guest}
  <GuestRoom code={decodeURIComponent(guest[1]).toLowerCase()} />
{:else if tv}
  <TvRoom token={tv[1]} />
{:else}
  <main class="not-found">
    <div class="card">
      <p class="badge">404</p>
      <h1>{$tr('notFound.title')}</h1>
      <p class="muted">{$tr('notFound.body')}</p>
      <a class="btn btn-primary" href="/">{$tr('common.home')}</a>
    </div>
    <div style="margin-top: 1rem;"><LanguageToggle /></div>
  </main>
{/if}

<style>
  .not-found {
    min-height: 100dvh;
    display: grid;
    place-content: center;
    justify-items: center;
    padding: 2rem;
    text-align: center;
  }
  .card {
    padding: 2rem;
    display: grid;
    gap: 0.75rem;
    justify-items: center;
    max-width: 420px;
  }
  .card h1 {
    font-size: 1.8rem;
  }
  .card .btn {
    margin-top: 0.5rem;
  }
</style>
