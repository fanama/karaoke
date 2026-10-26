<script lang="ts">
  import { tr } from '../../application/i18n';

  const cells = Array.from({ length: 49 }, (_, i) => (i * 7 + Math.floor(i / 7) * 3) % 5 < 2);
</script>

<div class="stage-panel" aria-hidden="true" data-testid="stage-preview">
  <div class="caption">
    <span>{$tr('welcome.stageCaption')}</span>
    <span class="caption-code">BLUE-MIC</span>
  </div>

  <div class="screen">
    <div class="screen-top">
      <span class="code">BLUE-MIC</span>
      <span class="live">LIVE</span>
    </div>

    <div class="now">
      <p class="meta">{$tr('tv.nowSinging')}</p>
      <p class="title">Bohemian Rhapsody</p>
      <p class="channel">Karaoke version</p>
    </div>

    <div class="screen-bottom">
      <p class="next">
        <span class="meta">{$tr('room.upNext')}</span>
        Dancing Queen
      </p>
      <div class="qr">
        {#each cells as on}
          <i class:on></i>
        {/each}
      </div>
    </div>
  </div>
</div>

<style>
  .stage-panel {
    border: 1px solid var(--border);
    border-radius: var(--radius-0);
  }

  .caption {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    height: 40px;
    padding: 0 var(--space-4);
    border-bottom: 1px solid var(--border);
    color: var(--muted);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    letter-spacing: 0.16em;
    text-transform: uppercase;
  }

  .caption-code {
    color: var(--text);
  }

  .screen {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    gap: var(--space-4);
    aspect-ratio: 16 / 10;
    max-height: 460px;
    padding: var(--space-4);
    background: var(--bg);
  }

  .screen-top,
  .screen-bottom {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-4);
  }

  .code {
    color: var(--accent);
    font-family: var(--mono);
    font-size: var(--text-lg);
    font-weight: 800;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }

  .live {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--accent);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    letter-spacing: 0.16em;
  }

  .live::before {
    content: "";
    width: 6px;
    height: 6px;
    background: var(--accent);
  }

  .meta {
    color: var(--muted);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    letter-spacing: 0.16em;
    text-transform: uppercase;
  }

  .title {
    margin-top: var(--space-2);
    font-size: clamp(1.25rem, 3vw, 2.25rem);
    font-weight: 800;
    letter-spacing: -0.03em;
    line-height: 0.95;
    text-transform: uppercase;
  }

  .channel {
    margin-top: var(--space-2);
    color: var(--muted);
    font-size: var(--text-xs);
  }

  .next {
    font-size: var(--text-sm);
    line-height: 1.3;
  }

  .next .meta {
    display: block;
    margin-bottom: 2px;
  }

  .qr {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: calc(var(--space-1) / 2);
    width: 72px;
    height: 72px;
    padding: calc(var(--space-1) + 2px);
    background: var(--paper);
    flex: none;
  }

  .qr i {
    background: transparent;
  }

  .qr i.on {
    background: var(--ink);
  }

  @media (min-width: 700px) {
    .screen {
      padding: var(--space-5);
    }

    .code {
      font-size: var(--text-2xl);
    }

    .qr {
      width: 88px;
      height: 88px;
    }
  }
</style>
