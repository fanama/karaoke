<script lang="ts">
  import { tr } from '../../application/i18n';
</script>

<section id="privacy" class="privacy paper" aria-labelledby="privacy-title">
  <h2 id="privacy-title">{$tr('welcome.privacyTitle')}</h2>
  <div>
    <p class="body">{$tr('welcome.privacyBody')}</p>
    <p class="note">{$tr('welcome.youtubeNote')}</p>
  </div>
</section>

<style>
  .privacy {
    display: grid;
    gap: var(--space-4);
    padding: var(--space-5);
  }

  h2 {
    font-size: var(--text-md);
    font-weight: 800;
    letter-spacing: -0.02em;
    text-transform: uppercase;
  }

  .body {
    max-width: 68ch;
    color: var(--ink);
    font-size: var(--text-sm);
    line-height: 1.6;
  }

  .note {
    margin-top: var(--space-4);
    padding-top: var(--space-3);
    border-top: 1px solid var(--ink-muted);
    max-width: 68ch;
    color: var(--ink-muted);
    font-size: var(--text-2xs);
    line-height: 1.5;
  }

  @media (min-width: 900px) {
    .privacy {
      grid-template-columns: 300px minmax(0, 1fr);
      align-items: start;
      padding: var(--space-6);
    }
  }
</style>
