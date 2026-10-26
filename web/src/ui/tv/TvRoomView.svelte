<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { tr } from '../../application/i18n';
  import Icon from '../common/Icon.svelte';
  import { createTvController } from '../../application/tv/controller';
  import { getAppContext } from '../../ports/context';
  import { computeElapsed, secondsUntil } from '../../domain/playback';
  import TvBottomBar from './TvBottomBar.svelte';
  import TvPlayer from './TvPlayer.svelte';
  import TvQueueDrawer from './TvQueueDrawer.svelte';
  import TvSkipBanner from './TvSkipBanner.svelte';
  import TvStageOverlays from './TvStageOverlays.svelte';
  import TvSoundPrompt from './TvSoundPrompt.svelte';
  import TvTopBar from './TvTopBar.svelte';

  export let token: string;

  const c = createTvController(token, getAppContext());
  const {
    code, pin, badLink, muted, playerBlocked, soundBlocked, queueOpen, busy, confirmClear,
    confirmEnd, dragId, snap, status, ended, toast, setPlayer, boot, playback, telemetry,
    toggleMute, turnSoundOn, onSoundBlocked, onPlayerBlocked, copyLink, newRoom, dispose,
    fullscreen, toggleLock, removeSong, clearQueue, rotatePin, endRoom, armEnd, dragStart,
    drop, cancelVote
  } = c;

  let player: TvPlayer | null = null;
  let nowTick = Date.now();
  let controlsVisible = true;
  let hideTimer: ReturnType<typeof setTimeout> | null = null;

  onMount(() => {
    const ticker = setInterval(() => (nowTick = Date.now()), 500);
    window.addEventListener('keydown', onKey);
    window.addEventListener('mousemove', wakeControls);
    window.addEventListener('touchstart', wakeControls);
    boot();
    return () => {
      clearInterval(ticker);
      window.removeEventListener('keydown', onKey);
      window.removeEventListener('mousemove', wakeControls);
      window.removeEventListener('touchstart', wakeControls);
      if (hideTimer) clearTimeout(hideTimer);
      dispose();
    };
  });

  $: setPlayer(player);
  $: current = $snap;
  $: queue = current?.queue ?? [];
  $: countdownLeft = current?.nowPlaying?.countdownEndsAt
    ? secondsUntil(current.nowPlaying.countdownEndsAt, nowTick)
    : 0;
  $: voteLeft = current?.skipVote?.expiresAt ? secondsUntil(current.skipVote.expiresAt, nowTick) : 0;
  $: elapsedNow = current ? computeElapsed(current, nowTick) : 0;
  $: if (current && (current.nowPlaying?.status !== 'playing' || $playerBlocked)) {
    controlsVisible = true;
    if (hideTimer) clearTimeout(hideTimer);
    hideTimer = null;
  } else if (current && !hideTimer) {
    wakeControls();
  }

  function wakeControls() {
    controlsVisible = true;
    if (hideTimer) clearTimeout(hideTimer);
    hideTimer = setTimeout(() => (controlsVisible = false), 3500);
  }

  function onKey(event: KeyboardEvent) {
    const target = event.target as HTMLElement;
    if (target?.tagName === 'INPUT' || target?.tagName === 'TEXTAREA' || !current) return;
    if (event.code === 'Space') {
      event.preventDefault();
      playback(current.nowPlaying?.status === 'playing' ? 'pause' : 'play');
    } else if (event.key.toLowerCase() === 'n') {
      event.preventDefault();
      playback('next');
    } else if (event.key.toLowerCase() === 'f') {
      event.preventDefault();
      fullscreen();
    } else if (event.key.toLowerCase() === 'q') {
      queueOpen.set(!$queueOpen);
    }
  }
</script>

<svelte:head>
  <title>{$code ? `${$code.toUpperCase()} · Fana Karaoke TV` : 'Fana Karaoke TV'}</title>
</svelte:head>

{#if $badLink}
  <main class="center-page">
    <div class="card end-card">
      <span class="badge">404</span>
      <h1>{$tr('tv.badLink')}</h1>
      <a class="btn btn-primary" href="/">{$tr('common.home')}</a>
    </div>
  </main>
{:else if $ended}
  <main class="center-page">
    <div class="card end-card">
      <span class="badge badge-hot">{$tr('app.name')}</span>
      <h1>{$tr('tv.endedTitle')}</h1>
      <p class="muted">{$tr('tv.endedBody')}</p>
      <button class="btn btn-primary" type="button" on:click={newRoom}>{$tr('common.createRoom')}</button>
      <a class="btn btn-ghost" href="/">{$tr('common.home')}</a>
    </div>
  </main>
{:else if !current}
  <main class="center-page">
    <span class="spinner large"></span>
    <p class="muted">{$tr('common.loading')}</p>
  </main>
{:else}
  <div class="tv">
    <TvTopBar
      code={current.room.code}
      pin={$pin}
      guestCount={current.room.guestCount}
      status={$status}
      onCopy={copyLink}
    />

    <div class="stage">
      {#if current.nowPlaying}
        <TvPlayer
          bind:this={player}
          videoId={current.nowPlaying.videoId}
          entryId={current.nowPlaying.entryId}
          generation={current.nowPlaying.generation}
          status={current.nowPlaying.status}
          elapsed={elapsedNow}
          muted={$muted}
          blocked={$playerBlocked}
          on:ended={(e) => telemetry('tvEnded', { entryId: e.detail.entryId, generation: e.detail.generation })}
          on:error={(e) => telemetry('tvError', { entryId: e.detail.entryId, generation: e.detail.generation, code: e.detail.code })}
          on:progress={(e) => telemetry('tvProgress', {
            entryId: e.detail.entryId,
            generation: e.detail.generation,
            positionSeconds: e.detail.position,
            durationSeconds: e.detail.duration
          })}
          on:blocked={(e) => onPlayerBlocked(e.detail)}
          on:soundblocked={(e) => onSoundBlocked(e.detail)}
        />
      {:else}
        <div class="stage-placeholder">
          <div class="stage-orb" aria-hidden="true"><Icon name="microphone" size={64} /></div>
        </div>
      {/if}
      {#if $soundBlocked && !$playerBlocked}
        <TvSoundPrompt onEnable={turnSoundOn} />
      {/if}
    </div>

    <TvStageOverlays
      snap={current}
      {countdownLeft}
      onPlay={() => playback('play')}
      onCancelCountdown={() => playback('cancelCountdown')}
    />

    {#if current.skipVote?.active}
      <TvSkipBanner vote={current.skipVote} {voteLeft} onCancel={cancelVote} />
    {/if}

    <TvBottomBar
      snap={current}
      {queue}
      {elapsedNow}
      muted={$muted}
      busy={$busy}
      hidden={!controlsVisible}
      onPlayback={playback}
      onToggleMute={toggleMute}
      onFullscreen={fullscreen}
      onToggleQueue={() => queueOpen.set(!$queueOpen)}
    />

    <TvQueueDrawer
      {queue}
      open={$queueOpen}
      manualOrder={current.room.manualOrder}
      queueLocked={current.room.queueLocked}
      confirmClear={$confirmClear}
      confirmEnd={$confirmEnd}
      dragId={$dragId}
      onClose={() => queueOpen.set(false)}
      onToggleLock={toggleLock}
      onRotatePin={rotatePin}
      onDragStart={dragStart}
      onDrop={drop}
      onRemove={removeSong}
      onClear={clearQueue}
      onArmClear={() => confirmClear.set(true)}
      onArmEnd={armEnd}
      onCancelClear={() => confirmClear.set(false)}
    />

    {#if $queueOpen}
      <button class="drawer-scrim" type="button" on:click={() => queueOpen.set(false)} aria-label={$tr('tv.closeQueue')}></button>
    {/if}

    {#if $toast}
      <div class="toast tv-toast" role="status" aria-live="polite">{$toast}</div>
    {/if}
  </div>
{/if}

<style>
  .tv {
    position: relative;
    display: flex;
    flex-direction: column;
    height: 100dvh;
    overflow: hidden;
    background: var(--bg);
  }

  .stage {
    position: relative;
    flex: 1;
    min-height: 0;
    background: var(--bg);
  }

  .stage-placeholder {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    background: var(--surface);
  }

  .stage-orb {
    display: grid;
    place-items: center;
    width: 180px;
    height: 180px;
    border: 1px solid var(--accent);
    color: var(--accent);
  }

  .drawer-scrim {
    position: absolute;
    inset: 0;
    z-index: 50;
    border: 0;
    background: rgb(11 11 12 / 0.72);
  }

  .tv-toast {
    bottom: calc(var(--space-9) + var(--space-8));
  }
</style>
