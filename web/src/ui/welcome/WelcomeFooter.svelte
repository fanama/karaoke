<script lang="ts">
  import { tr } from '../../application/i18n';
</script>

<footer>
  <div class="footer-left">
    <span>Fana Karaoke</span>
    <span class="dim">{$tr('app.tagline')}</span>
  </div>
  <span class="credit">
    {$tr('app.credit')}
    <a href="https://github.com/fanama" target="_blank" rel="noopener noreferrer">github.com/fanama</a>
  </span>
</footer>

<style>
  footer {
    position: relative;
    z-index: 1;
    width: 100%;
    max-width: 1180px;
    margin: 0 auto;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 0.6rem 1rem;
    padding: 1.5rem 0;
    border-top: 1px solid var(--border);
    color: var(--muted);
    font-size: 0.8rem;
  }

  .footer-left {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.35rem 1rem;
  }

  .credit {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
  }

  .credit a {
    color: var(--cyan);
    font-weight: 700;
  }

  .credit a:hover {
    text-decoration: underline;
  }

  @media (max-width: 420px) {
    footer {
      flex-direction: column;
      align-items: flex-start;
    }
  }
</style>
