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
    <p class="tiny dim">{$tr('tv.pinHelp')}</p>
  </div>
</div>

<style>
  .join-panel {
    display: flex;
    gap: 0.85rem;
    align-items: center;
    padding: 0.7rem;
    border-radius: 16px;
    border: 1px solid rgb(74 42 114 / 0.75);
    background: rgb(8 5 15 / 0.7);
    backdrop-filter: blur(10px);
    box-shadow: 0 10px 40px rgb(0 0 0 / 0.35);
  }

  .qr-box {
    width: 112px;
    height: 112px;
    border-radius: 8px;
    overflow: hidden;
    background: white;
    flex: none;
  }

  .qr-box img {
    display: block;
    width: 100%;
    height: 100%;
  }

  .join-copy {
    display: grid;
    gap: 0.15rem;
    align-content: center;
  }

  .join-label {
    font-size: 0.66rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    color: var(--cyan);
  }

  .join-code {
    font-size: 1.05rem;
    font-weight: 800;
    letter-spacing: 0.08em;
  }

  .pin-line {
    display: flex;
    align-items: baseline;
    gap: 0.45rem;
    margin-top: 0.28rem;
  }

  .pin-line .label {
    font-size: 0.62rem;
  }

  .pin {
    font-size: 1.15rem;
    letter-spacing: 0.14em;
    color: var(--magenta);
  }

  .join-copy .tiny {
    max-width: 20ch;
    line-height: 1.35;
  }

  @media (max-width: 900px) {
    .join-panel {
      padding: 0.5rem;
    }
    .qr-box {
      width: 84px;
      height: 84px;
    }
    .join-copy .tiny {
      display: none;
    }
  }

  @media (max-width: 600px) {
    .join-panel {
      flex-direction: column;
      gap: 0.4rem;
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
