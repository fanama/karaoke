<script lang="ts">
  import { onMount } from 'svelte';
  import { tr } from './application/i18n';
  import { setAppContext, type AppDependencies } from './ports/context';
  import LanguageToggle from './ui/common/LanguageToggle.svelte';
  import GuestRoomView from './ui/guest/GuestRoomView.svelte';
  import TvRoomView from './ui/tv/TvRoomView.svelte';
  import WelcomeView from './ui/welcome/WelcomeView.svelte';

  export let deps: AppDependencies;

  setAppContext(deps);

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
  <WelcomeView />
{:else if guest}
  <GuestRoomView code={decodeURIComponent(guest[1]).toLowerCase()} />
{:else if tv}
  <TvRoomView token={tv[1]} />
{:else}
  <main class="center-page">
    <div class="card end-card">
      <p class="badge">404</p>
      <h1>{$tr('notFound.title')}</h1>
      <p class="muted">{$tr('notFound.body')}</p>
      <a class="btn btn-primary" href="/">{$tr('common.home')}</a>
    </div>
    <div style="margin-top: 1rem;"><LanguageToggle /></div>
  </main>
{/if}
