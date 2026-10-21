<script lang="ts">
  import { lang, setLang } from '../i18n';
</script>

<div class="lang-toggle" role="group" aria-label="Language">
  <button type="button" class:active={$lang === 'en'} on:click={() => setLang('en')}>EN</button>
  <button type="button" class:active={$lang === 'fr'} on:click={() => setLang('fr')}>FR</button>
</div>
