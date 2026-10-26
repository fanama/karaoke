<script lang="ts">
  import { tr } from '../../application/i18n';
</script>

<footer>
  <span class="brand">Fana Karaoke</span>
  <span class="tagline">{$tr('app.tagline')}</span>
  <span class="credit">
    {$tr('app.credit')}
    <a href="https://github.com/fanama" target="_blank" rel="noopener noreferrer">github.com/fanama</a>
  </span>
</footer>

<style>
  footer {
    display: grid;
    gap: var(--space-2);
    margin-top: var(--space-7);
    padding-top: var(--space-4);
    border-top: 1px solid var(--border);
    color: var(--muted);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }

  .brand {
    color: var(--text);
  }

  .credit {
    display: inline-flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: var(--space-2);
  }

  .credit a {
    display: inline-flex;
    align-items: center;
    min-height: 44px;
    color: var(--text);
    text-decoration: underline;
    text-decoration-thickness: 1px;
    text-underline-offset: 3px;
  }

  .credit a:hover {
    color: var(--accent);
  }

  @media (min-width: 900px) {
    footer {
      grid-template-columns: auto 1fr auto;
      align-items: baseline;
      gap: var(--space-5);
    }

    .tagline {
      text-align: center;
    }
  }
</style>
