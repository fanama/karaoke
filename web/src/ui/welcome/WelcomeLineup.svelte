<script lang="ts">
  import { tr } from '../../application/i18n';

  const rows = [
    { number: '01', title: 'welcome.feature1Title', body: 'welcome.feature1Body' },
    { number: '02', title: 'welcome.feature2Title', body: 'welcome.feature2Body' },
    { number: '03', title: 'welcome.feature3Title', body: 'welcome.feature3Body' }
  ];
</script>

<section class="lineup" aria-labelledby="lineup-title" data-testid="lineup">
  <h2 id="lineup-title">{$tr('welcome.lineupTitle')}</h2>

  <ol>
    <li class="lead">
      <span class="num">{rows[0].number}</span>
      <h3>{$tr(rows[0].title)}</h3>
      <p>{$tr(rows[0].body)}</p>
    </li>
    <li class="half">
      <span class="num">{rows[1].number}</span>
      <div>
        <h3>{$tr(rows[1].title)}</h3>
        <p>{$tr(rows[1].body)}</p>
      </div>
    </li>
    <li class="half">
      <span class="num">{rows[2].number}</span>
      <div>
        <h3>{$tr(rows[2].title)}</h3>
        <p>{$tr(rows[2].body)}</p>
      </div>
    </li>
  </ol>
</section>

<style>
  .lineup h2 {
    padding-top: var(--space-3);
    border-top: 2px solid var(--accent);
    font-size: var(--text-lg);
    font-weight: 800;
    letter-spacing: -0.03em;
    line-height: 1.05;
    text-transform: uppercase;
    text-wrap: balance;
  }

  ol {
    display: grid;
    margin: var(--space-5) 0 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: grid;
    grid-template-columns: 40px minmax(0, 1fr);
    gap: var(--space-2) var(--space-3);
    padding: var(--space-4) 0;
    border-top: 1px solid var(--border);
  }

  .num {
    color: var(--accent);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    letter-spacing: 0.1em;
  }

  h3 {
    font-size: var(--text-md);
    font-weight: 800;
    letter-spacing: -0.02em;
    text-transform: uppercase;
  }

  p {
    grid-column: 2;
    max-width: 46ch;
    color: var(--muted);
    font-size: var(--text-sm);
    line-height: 1.55;
  }

  .lead p {
    grid-column: 2;
  }

  @media (min-width: 900px) {
    ol {
      grid-template-columns: 1fr 1fr;
    }

    li {
      padding: var(--space-5) 0;
    }

    .lead {
      grid-column: 1 / -1;
      grid-template-columns: 56px 18ch minmax(0, 1fr);
      align-items: baseline;
      gap: var(--space-5);
    }

    .lead h3 {
      font-size: var(--text-lg);
    }

    .lead p {
      grid-column: 3;
    }

    .half {
      grid-template-columns: 40px minmax(0, 1fr);
      gap: var(--space-4);
    }

    .half + .half {
      padding-left: var(--space-5);
      border-left: 1px solid var(--border);
    }
  }
</style>
