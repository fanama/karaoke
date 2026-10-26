<script lang="ts">
  import { tr } from '../../application/i18n';

  const steps = [
    { number: '01', body: 'welcome.howStep1' },
    { number: '02', body: 'welcome.howStep2' },
    { number: '03', body: 'welcome.howStep3' }
  ];
</script>

<section id="how" class="programme paper" aria-labelledby="how-title">
  <h2 id="how-title">{$tr('welcome.howTitle')}</h2>
  <ol>
    {#each steps as step}
      <li>
        <span class="step">{step.number}</span>
        <p>{$tr(step.body)}</p>
      </li>
    {/each}
  </ol>
</section>

<style>
  .programme {
    padding: var(--space-5);
  }

  h2 {
    font-size: var(--text-md);
    font-weight: 800;
    letter-spacing: -0.02em;
    text-transform: uppercase;
  }

  ol {
    display: grid;
    margin: var(--space-5) 0 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: grid;
    grid-template-columns: 32px minmax(0, 1fr);
    gap: var(--space-3);
    padding: var(--space-5) 0;
  }

  li:first-child {
    padding-top: 0;
  }

  li:last-child {
    padding-bottom: 0;
  }

  .step {
    color: var(--ink-muted);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    letter-spacing: 0.14em;
  }

  p {
    color: var(--ink);
    font-size: var(--text-sm);
    line-height: 1.5;
  }

  @media (min-width: 900px) {
    .programme {
      padding: var(--space-6);
    }

    ol {
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: var(--space-6);
    }

    li {
      grid-template-columns: 1fr;
      gap: var(--space-3);
      padding: 0;
    }
  }
</style>
