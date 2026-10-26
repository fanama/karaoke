<script lang="ts">
  import { tr } from '../../application/i18n';

  export let code: string;
  export let pin: string;
</script>

<div class="join-panel">
  <div class="qr-box">
    <img src={`/api/rooms/${code}/qr.svg`} alt={$tr('room.scan')} width="112" height="112" />
  </div>
  <div class="join-copy">
    <p class="join-label">{$tr('room.scan')}</p>
    <p class="join-code mono">{code.toUpperCase()}</p>
    <p class="pin-line">
      <span class="label">{$tr('room.pin')}</span>
      <strong class="pin mono">{pin}</strong>
    </p>
    <p class="pin-help">{$tr('tv.pinHelp')}</p>
  </div>
</div>

<style>
  .join-panel {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2);
    background: var(--paper);
    border-radius: var(--radius-0);
    color: var(--ink);
  }

  .qr-box {
    width: 112px;
    height: 112px;
    flex: none;
    overflow: hidden;
    background: #fff;
  }

  .qr-box img {
    display: block;
    width: 100%;
    height: 100%;
  }

  .join-copy {
    display: grid;
    align-content: center;
    gap: var(--space-1);
  }

  .join-label {
    color: var(--ink-muted);
    font-family: var(--mono);
    font-size: var(--text-2xs);
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }

  .join-code {
    font-size: var(--text-md);
    font-weight: 800;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .pin-line {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
  }

  .pin-line .label {
    color: var(--ink-muted);
    font-size: var(--text-2xs);
  }

  .pin {
    font-size: var(--text-md);
    letter-spacing: 0.14em;
  }

  .pin-help {
    max-width: 22ch;
    color: var(--ink-muted);
    font-size: var(--text-2xs);
    line-height: 1.4;
  }

  @media (max-width: 900px) {
    .qr-box {
      width: 84px;
      height: 84px;
    }

    .pin-help {
      display: none;
    }
  }

  @media (max-width: 600px) {
    .join-panel {
      flex-direction: column;
      gap: var(--space-1);
    }

    .qr-box {
      width: 70px;
      height: 70px;
    }

    .join-copy {
      display: none;
    }
  }
</style>
