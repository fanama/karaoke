<script lang="ts">
  const bars = [14, 30, 22, 44, 60, 36, 72, 48, 26, 54, 34, 18, 40, 64, 28, 20, 38, 50];
</script>

<div class="stage-card" aria-hidden="true">
  <div class="stage-screen">
    <div class="screen-top">
      <span class="screen-code">BLUE-MIC</span>
      <span class="screen-live"><i></i>LIVE</span>
    </div>
    <div class="screen-title">
      <span class="disc">♪</span>
      <div>
        <strong>Bohemian Rhapsody</strong>
        <small>Karaoke version</small>
      </div>
    </div>
    <div class="wave">
      {#each bars as h}
        <span style={`height:${h}%`}></span>
      {/each}
    </div>
    <div class="screen-bottom">
      <span class="badge badge-hot">ALEX</span>
      <span class="screen-next">Next · Dancing Queen</span>
    </div>
    <div class="qr-sim">
      {#each Array(49) as _, i}
        <i class:on={(i * 7 + Math.floor(i / 7) * 3) % 5 < 2}></i>
      {/each}
    </div>
  </div>
  <div class="stage-glow"></div>
</div>

<style>
  .stage-card {
    position: relative;
    perspective: 1200px;
  }

  .stage-screen {
    position: relative;
    aspect-ratio: 16 / 10;
    padding: 1.3rem;
    border-radius: 24px;
    border: 1px solid #542e83;
    background:
      radial-gradient(circle at 80% 0%, rgb(255 45 149 / 0.22), transparent 45%),
      radial-gradient(circle at 0% 100%, rgb(34 211 238 / 0.16), transparent 45%),
      #0c0716;
    box-shadow: 0 30px 80px rgb(0 0 0 / 0.55), 0 0 50px rgb(255 45 149 / 0.12);
    transform: rotateY(-7deg) rotateX(3deg);
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    overflow: hidden;
  }

  .screen-top,
  .screen-bottom {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .screen-code {
    font-family: var(--mono);
    font-size: clamp(1.2rem, 2.4vw, 2rem);
    font-weight: 800;
    letter-spacing: 0.12em;
  }

  .screen-live {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    color: var(--lime);
    font-size: 0.65rem;
    font-weight: 800;
    letter-spacing: 0.16em;
  }

  .screen-live i {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--lime);
    box-shadow: 0 0 10px var(--lime);
  }

  .screen-title {
    display: flex;
    align-items: center;
    gap: 0.8rem;
  }

  .screen-title strong {
    display: block;
    font-size: clamp(1.05rem, 2vw, 1.5rem);
  }

  .screen-title small {
    color: var(--muted);
  }

  .disc {
    display: grid;
    place-items: center;
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--magenta), var(--violet));
    font-size: 1.4rem;
    color: #1b0010;
  }

  .wave {
    display: flex;
    align-items: center;
    gap: 4px;
    height: 76px;
    opacity: 0.9;
  }

  .wave span {
    flex: 1;
    border-radius: 4px;
    background: linear-gradient(to top, var(--magenta), var(--cyan));
    min-height: 4px;
  }

  .qr-sim {
    position: absolute;
    right: 1.2rem;
    bottom: 2.7rem;
    width: 72px;
    height: 72px;
    padding: 6px;
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 2px;
    background: white;
    border-radius: 6px;
    opacity: 0.9;
  }

  .qr-sim i {
    background: transparent;
    border-radius: 1px;
  }

  .qr-sim i.on {
    background: #08050f;
  }

  .stage-glow {
    position: absolute;
    inset: auto 10% -30px;
    height: 70px;
    background: var(--magenta);
    filter: blur(48px);
    opacity: 0.4;
    z-index: -1;
  }

  @media (max-width: 900px) {
    .stage-card {
      max-width: 540px;
    }
  }
</style>
