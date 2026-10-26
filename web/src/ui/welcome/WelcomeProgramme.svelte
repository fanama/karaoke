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
    border-top: 2px solid var(--accent);
  }

  h2 {
    font-size: var(--text-md);
    font-weight: 800;
    letter-spacing: -0.02em;
    text-transform: uppercase;
  }

  ol {
    display: grid;
    margin: var(--space-4) 0 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: grid;
    grid-template-columns: 32px minmax(0, 1fr);
    gap: var(--space-2);
    padding: var(--space-4) 0;
    border-top: 1px solid var(--ink-muted);
  }

  li:first-child {
    border-top: 0;
    padding-top: 0;
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
    }

    li {
      grid-template-columns: 1fr;
      gap: var(--space-2);
      padding: 0 var(--space-4);
      border-top: 0;
      border-left: 1px solid var(--ink-muted);
    }

    li:first-child {
      padding-left: 0;
      border-left: 0;
    }

    li:last-child {
      padding-right: 0;
    }
  }
</style>
