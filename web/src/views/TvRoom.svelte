<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { api, ApiError } from '../lib/api';
  import ConnectionDot from '../lib/components/ConnectionDot.svelte';
  import LanguageToggle from '../lib/components/LanguageToggle.svelte';
  import { errorText, noticeText, tr } from '../lib/i18n';
  import { formatDuration, secondsUntil } from '../lib/format';
  import type { Snapshot } from '../lib/types';
  import { connectRoom, type ConnectionStatus, type RoomSocket } from '../lib/ws';
  import Player from '../features/tv/Player.svelte';

  export let token: string;

  let session: { code: string; pin: string } | null = null;
  let code = '';
  let pin = '';
  let snap: Snapshot | null = null;
  let status: ConnectionStatus = 'connecting';
  let ended = false;
  let badLink = false;
  let socket: RoomSocket | null = null;
  let player: Player | null = null;
  let nowTick = Date.now();
  let controlsVisible = true;
  let hideTimer: ReturnType<typeof setTimeout> | null = null;
  let queueOpen = false;
  let muted = false;
  let toast = '';
  let toastTimer: ReturnType<typeof setTimeout> | null = null;
  let lastNoticeId = '';
  let seenFirst = false;
  let confirmClear = false;
  let confirmEnd = false;
  let confirmEndTimer: ReturnType<typeof setTimeout> | null = null;
  let roomCheckTimer: ReturnType<typeof setTimeout> | null = null;
  let dragId = '';
  let busy = false;
  let playerBlocked = false;
  let soundBlocked = false;

  let countdownLeft = 0;
  let voteLeft = 0;
  let elapsedNow = 0;
  let queue = [] as Snapshot['queue'];

  $: queue = snap?.queue ?? [];
  $: if (snap) {
    const _tick = nowTick;
    countdownLeft = snap.nowPlaying?.countdownEndsAt
      ? secondsUntil(snap.nowPlaying.countdownEndsAt)
      : 0;
    voteLeft = snap.skipVote?.expiresAt ? secondsUntil(snap.skipVote.expiresAt) : 0;
    elapsedNow = snap.nowPlaying
      ? snap.nowPlaying.elapsedSeconds +
        (snap.nowPlaying.status === 'playing'
          ? Math.max(0, (Date.now() - new Date(snap.serverTime).getTime()) / 1000)
          : 0)
      : 0;
  }

  onMount(() => {
    const ticker = setInterval(() => (nowTick = Date.now()), 500);
    boot();
    return () => {
      clearInterval(ticker);
      socket?.close();
      if (toastTimer) clearTimeout(toastTimer);
      if (hideTimer) clearTimeout(hideTimer);
      if (confirmEndTimer) clearTimeout(confirmEndTimer);
      if (roomCheckTimer) clearTimeout(roomCheckTimer);
    };
  });

  async function boot() {
    if (!token) {
      badLink = true;
      return;
    }
    try {
      session = await api.tvSession();
      code = session.code;
      pin = session.pin;
      snap = await api.getRoom(code);
      seenFirst = true;
      lastNoticeId = snap.notices[0]?.id ?? '';
      connect();
      document.title = `${snap.room.code.toUpperCase()} · Fana Karaoke`;
    } catch {
      badLink = true;
    }
  }

  function connect() {
    socket?.close();
    socket = connectRoom(code, 'tv', {
      onStatus: (s) => {
        status = s;
        if (s === 'offline') scheduleRoomCheck();
        else if (roomCheckTimer) {
          clearTimeout(roomCheckTimer);
          roomCheckTimer = null;
        }
      },
      onSnapshot: (s) => {
        snap = s;
        handleNotice(s);
      },
      onEnded: () => {
        ended = true;
        socket?.close();
      }
    });
  }

  function scheduleRoomCheck() {
    if (roomCheckTimer) return;
    roomCheckTimer = setTimeout(async () => {
      roomCheckTimer = null;
      try {
        await api.getRoom(code);
      } catch (err) {
        if (err instanceof ApiError && (err.code === 'room_not_found' || err.status === 404)) {
          ended = true;
          socket?.close();
        }
      }
    }, 8000);
  }

  function handleNotice(s: Snapshot) {
    const latest = s.notices[0];
    if (!seenFirst || !latest || latest.id === lastNoticeId) return;
    lastNoticeId = latest.id;
    showToast(noticeText($tr, latest.code, latest.params));
  }

  function showToast(text: string) {
    if (!text) return;
    toast = text;
    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(() => (toast = ''), 4000);
  }

  function message(err: unknown) {
    return err instanceof ApiError ? errorText($tr, err.code) : $tr('error.generic');
  }

  async function playback(action: 'play' | 'pause' | 'next' | 'cancelCountdown') {
    if (busy) return;
    busy = true;
    if (action !== 'pause') playerBlocked = false;
    try {
      snap = await api.playback(code, action);
      if (action === 'play') player?.forcePlay();
    } catch (err) {
      showToast(message(err));
    } finally {
      busy = false;
    }
  }

  function telemetryEnded(detail: { entryId: string; generation: number }) {
    socket?.send({ type: 'tvEnded', entryId: detail.entryId, generation: detail.generation });
  }

  function telemetryError(detail: { entryId: string; generation: number; code: string }) {
    socket?.send({
      type: 'tvError',
      entryId: detail.entryId,
      generation: detail.generation,
      code: detail.code
    });
  }

  function telemetryProgress(detail: {
    entryId: string;
    generation: number;
    position: number;
    duration: number;
  }) {
    socket?.send({
      type: 'tvProgress',
      entryId: detail.entryId,
      generation: detail.generation,
      positionSeconds: detail.position,
      durationSeconds: detail.duration
    });
  }

  async function toggleLock() {
    if (!snap) return;
    try {
      snap = await api.lockQueue(code, !snap.room.queueLocked);
    } catch (err) {
      showToast(message(err));
    }
  }

  async function removeSong(id: string) {
    try {
      snap = await api.removeSong(code, id);
    } catch (err) {
      showToast(message(err));
    }
  }

  async function clearQueue() {
    try {
      snap = await api.clearQueue(code);
      confirmClear = false;
    } catch (err) {
      showToast(message(err));
    }
  }

  async function rotatePin() {
    try {
      const res = await api.rotatePIN(code);
      pin = res.pin;
      showToast($tr('room.pinChanged'));
    } catch (err) {
      showToast(message(err));
    }
  }

  async function endRoom() {
    try {
      await api.endRoom(code);
      ended = true;
      socket?.close();
    } catch (err) {
      showToast(message(err));
    }
  }

  function armEnd() {
    if (confirmEnd) {
      endRoom();
      return;
    }
    confirmEnd = true;
    if (confirmEndTimer) clearTimeout(confirmEndTimer);
    confirmEndTimer = setTimeout(() => (confirmEnd = false), 3500);
  }

  async function reorder(ids: string[]) {
    try {
      snap = await api.reorder(code, ids);
    } catch (err) {
      showToast(message(err));
    }
  }

  function onDragStart(event: DragEvent, id: string) {
    dragId = id;
    event.dataTransfer?.setData('text/plain', id);
    if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move';
  }

  function onDrop(event: DragEvent, targetId: string) {
    event.preventDefault();
    if (!dragId || dragId === targetId) return;
    const ids = queue.map((q) => q.id);
    const from = ids.indexOf(dragId);
    const to = ids.indexOf(targetId);
    if (from < 0 || to < 0) return;
    ids.splice(from, 1);
    ids.splice(to, 0, dragId);
    dragId = '';
    reorder(ids);
  }

  async function cancelVote() {
    try {
      snap = await api.cancelSkipVote(code);
    } catch (err) {
      showToast(message(err));
    }
  }

  async function copyLink() {
    try {
      await navigator.clipboard.writeText(location.href);
      showToast($tr('tv.tvLinkCopied'));
    } catch {
      showToast(location.href);
    }
  }

  function fullscreen() {
    if (document.fullscreenElement) {
      document.exitFullscreen?.();
    } else {
      document.documentElement.requestFullscreen?.();
    }
  }

  function toggleMute() {
    muted = !muted;
    if (!muted) soundBlocked = false;
    player?.setMuted(muted);
  }

  function turnSoundOn() {
    soundBlocked = false;
    muted = false;
    player?.setMuted(false);
  }

  function onKey(event: KeyboardEvent) {
    const target = event.target as HTMLElement;
    if (target?.tagName === 'INPUT' || target?.tagName === 'TEXTAREA') return;
    if (!snap) return;
    if (event.code === 'Space') {
      event.preventDefault();
      playback(snap.nowPlaying?.status === 'playing' ? 'pause' : 'play');
    } else if (event.key.toLowerCase() === 'n') {
      event.preventDefault();
      playback('next');
    } else if (event.key.toLowerCase() === 'f') {
      event.preventDefault();
      fullscreen();
    } else if (event.key.toLowerCase() === 'q') {
      queueOpen = !queueOpen;
    }
  }

  function wakeControls() {
    controlsVisible = true;
    if (hideTimer) clearTimeout(hideTimer);
    hideTimer = setTimeout(() => (controlsVisible = false), 3500);
  }

  $: if (snap) {
    if (snap.nowPlaying?.status !== 'playing' || playerBlocked) controlsVisible = true;
    else if (!hideTimer) wakeControls();
  }

  function onSoundBlocked(event: CustomEvent<boolean>) {
    soundBlocked = event.detail;
    // Keep the toolbar icon honest: playback is silent until someone taps.
    muted = event.detail;
  }

  function onPlayerBlocked(event: CustomEvent<boolean>) {
    playerBlocked = event.detail;
    if (playerBlocked) {
      controlsVisible = true;
      if (hideTimer) clearTimeout(hideTimer);
      hideTimer = null;
    }
  }

  async function newRoom() {
    try {
      const room = await api.createRoom();
      location.assign(room.tvUrl);
    } catch (err) {
      showToast(message(err));
    }
  }
</script>

<svelte:head>
  <title>{code ? `${code.toUpperCase()} · Fana Karaoke TV` : 'Fana Karaoke TV'}</title>
</svelte:head>
<svelte:window on:keydown={onKey} on:mousemove={wakeControls} on:touchstart={wakeControls} />

{#if badLink}
  <main class="center-page">
    <div class="card end-card">
      <span class="badge">404</span>
      <h1>{$tr('tv.badLink')}</h1>
      <a class="btn btn-primary" href="/">{$tr('common.home')}</a>
    </div>
  </main>
{:else if ended}
  <main class="center-page">
    <div class="card end-card">
      <span class="badge badge-hot">{$tr('app.name')}</span>
      <h1>{$tr('tv.endedTitle')}</h1>
      <p class="muted">{$tr('tv.endedBody')}</p>
      <button class="btn btn-primary" type="button" on:click={newRoom}>{$tr('common.createRoom')}</button>
      <a class="btn btn-ghost" href="/">{$tr('common.home')}</a>
    </div>
  </main>
{:else if !snap}
  <main class="center-page">
    <span class="spinner large"></span>
    <p class="muted">{$tr('common.loading')}</p>
  </main>
{:else}
  <div class="tv" class:controls-hidden={!controlsVisible}>
    <div class="stage">
      {#if snap.nowPlaying}
        <Player
          bind:this={player}
          videoId={snap.nowPlaying.videoId}
          entryId={snap.nowPlaying.entryId}
          generation={snap.nowPlaying.generation}
          status={snap.nowPlaying.status}
          elapsed={elapsedNow}
          {muted}
          on:ended={(e) => telemetryEnded(e.detail)}
          on:error={(e) => telemetryError(e.detail)}
          on:progress={(e) => telemetryProgress(e.detail)}
          on:blocked={onPlayerBlocked}
          on:soundblocked={(e) => onSoundBlocked(e)}
        />
      {:else}
        <div class="stage-placeholder">
          <div class="stage-orb">♪</div>
        </div>
      {/if}
      <div class="vignette"></div>
      {#if soundBlocked && !playerBlocked}
        <button class="sound-prompt" type="button" on:click={turnSoundOn}>
          <span class="sound-prompt-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M4 9v6h4l5 4V5L8 9z"/>
              <path d="m17 9 4 6M21 9l-4 6"/>
            </svg>
          </span>
          <span class="sound-prompt-copy">
            <strong>{$tr('tv.soundOff')}</strong>
            <small>{$tr('tv.soundOffBody')}</small>
          </span>
        </button>
      {/if}
    </div>

    <header class="tv-top">
      <div class="room-mark">
        <p class="label">{$tr('room.code')}</p>
        <h1 class="tv-code">{snap.room.code.toUpperCase()}</h1>
        <div class="row">
          <ConnectionDot {status} />
          <span class="tv-stat">{$tr('tv.guestCount', { count: snap.room.guestCount })}</span>
        </div>
      </div>

      <div class="join-panel">
        <div class="qr-box">
          <img src={`/api/rooms/${snap.room.code}/qr.svg`} alt={$tr('room.scan')} width="112" height="112" />
        </div>
        <div class="join-copy">
          <p class="join-label">{$tr('room.scan')}</p>
          <p class="join-code mono">{snap.room.code.toUpperCase()}</p>
          <p class="pin-line">
            <span class="label">{$tr('room.pin')}</span>
            <strong class="pin mono">{pin}</strong>
          </p>
          <p class="tiny dim">{$tr('tv.pinHelp')}</p>
        </div>
      </div>

      <div class="tv-tools">
        <LanguageToggle />
        <button class="tool-btn" type="button" on:click={copyLink} title={$tr('room.copyTvLink')}>
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7 0l2-2a5 5 0 0 0-7-7l-1 1"/><path d="M14 11a5 5 0 0 0-7 0l-2 2a5 5 0 0 0 7 7l1-1"/></svg>
        </button>
        <a class="tool-btn" href={`/r/${snap.room.code}`} title={$tr('room.viewGuest')}>
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><rect x="7" y="2" width="10" height="20" rx="2"/><path d="M11 18h2"/></svg>
        </a>
      </div>
    </header>

    {#if !snap.room.tvStarted}
      <div class="stage-overlay start-overlay">
        <div class="start-card">
          <div class="start-icon" aria-hidden="true">🎤</div>
          <h2>{$tr('tv.start')}</h2>
          <p>{$tr('tv.startBody')}</p>
          <button class="btn btn-primary btn-lg" type="button" on:click={() => playback('play')}>
            {$tr('tv.start')}
          </button>
        </div>
      </div>
    {:else if !snap.nowPlaying}
      <div class="stage-overlay idle-overlay">
        <div class="idle-card">
          <div class="idle-note">♪</div>
          <h2>{$tr('tv.idleTitle')}</h2>
          <p>{$tr('tv.idleBody')}</p>
          <span class="badge badge-cyan">{$tr('tv.addSongs')}</span>
        </div>
      </div>
    {:else if snap.nowPlaying.status === 'countdown'}
      <div class="stage-overlay countdown-overlay">
        <div class="countdown-card">
          <p class="countdown-label">{$tr('tv.countdownNext')}</p>
          <h2 class="clamp-2">{snap.nowPlaying.title}</h2>
          <p class="countdown-singer">{$tr('tv.nowSinging')}: <strong>{snap.nowPlaying.nickname}</strong></p>
          <div class="countdown-number">{Math.max(1, countdownLeft)}</div>
          <button class="btn btn-secondary" type="button" on:click={() => playback('cancelCountdown')}>
            {$tr('tv.cancel')}
          </button>
        </div>
      </div>
    {/if}

    {#if snap.skipVote?.active}
      <div class="skip-banner">
        <div class="skip-copy">
          <strong>{$tr('tv.skipVoteTitle')}</strong>
          <span>{$tr('tv.voteProgress', { yes: snap.skipVote.yes, needed: snap.skipVote.threshold })}</span>
        </div>
        <div class="skip-bar">
          <span style={`width:${Math.min(100, (snap.skipVote.yes / Math.max(1, snap.skipVote.threshold)) * 100)}%`}></span>
        </div>
        <span class="skip-time mono">{voteLeft}s</span>
        <button class="btn btn-ghost btn-sm" type="button" on:click={cancelVote}>{$tr('tv.cancelVote')}</button>
      </div>
    {/if}

    <div class="tv-bottom">
      <div class="now-summary">
        {#if snap.nowPlaying}
          <img class="now-thumb" src={snap.nowPlaying.thumbnailUrl} alt="" />
          <div class="now-meta">
            <p class="now-title clamp-2">{snap.nowPlaying.title}</p>
            <p class="tiny muted truncate">
              {$tr('tv.nowSinging')}: <strong>{snap.nowPlaying.nickname}</strong>
              {#if snap.nowPlaying.channel} · {snap.nowPlaying.channel}{/if}
            </p>
            <div class="timeline">
              <span style={`width:${snap.nowPlaying.durationSeconds ? Math.min(100, (elapsedNow / snap.nowPlaying.durationSeconds) * 100) : 0}%`}></span>
            </div>
            <div class="spread tiny dim">
              <span>{formatDuration(elapsedNow)}</span>
              <span>{formatDuration(snap.nowPlaying.durationSeconds)}</span>
            </div>
          </div>
        {:else}
          <div class="now-meta">
            <p class="now-title">{$tr('tv.noSong')}</p>
            <p class="tiny muted">{$tr('tv.addSongs')}</p>
          </div>
        {/if}
      </div>

      <div class="playback-controls">
        {#if snap.nowPlaying?.status === 'playing'}
          <button class="ctrl" type="button" on:click={() => playback('pause')} disabled={busy} title={$tr('tv.pause')}>
            <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor"><rect x="6" y="5" width="4" height="14" rx="1"/><rect x="14" y="5" width="4" height="14" rx="1"/></svg>
          </button>
        {:else}
          <button class="ctrl primary" type="button" on:click={() => playback('play')} disabled={busy} title={$tr('tv.play')}>
            <svg viewBox="0 0 24 24" width="26" height="26" fill="currentColor"><path d="M7 5v14l12-7z"/></svg>
          </button>
        {/if}
        <button class="ctrl" type="button" on:click={() => playback('next')} disabled={busy} title={$tr('tv.next')}>
          <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor"><path d="M6 5v14l9-7z"/><rect x="16" y="5" width="2" height="14" rx="1"/></svg>
        </button>
        <button class="ctrl" type="button" on:click={toggleMute} title={$tr(muted ? 'tv.unmute' : 'tv.mute')}>
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 9v6h4l5 4V5L8 9z"/>
            {#if muted}<path d="m17 9 4 6M21 9l-4 6"/>{:else}<path d="M17 8a6 6 0 0 1 0 8"/>{/if}
          </svg>
        </button>
        <button class="ctrl" type="button" on:click={fullscreen} title={$tr('tv.fullscreen')}>
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5"/></svg>
        </button>
        <button class="ctrl queue-toggle" type="button" on:click={() => (queueOpen = !queueOpen)} title={$tr('tv.openQueue')}>
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01"/></svg>
          <span class="queue-badge">{snap.queue.length}</span>
        </button>
      </div>

      <div class="next-summary">
        {#if queue[0]}
          <p class="label">{$tr('room.upNext')}</p>
          <div class="next-row">
            <img src={queue[0].thumbnailUrl} alt="" />
            <div class="grow">
              <p class="clamp-2 next-title">{queue[0].title}</p>
              <p class="tiny dim">{queue[0].nickname} · {formatDuration(queue[0].durationSeconds)}</p>
            </div>
            <span class="badge">{queue[0].score} ↑</span>
          </div>
        {:else}
          <p class="label">{$tr('room.upNext')}</p>
          <p class="small dim">{$tr('tv.queueEmpty')}</p>
        {/if}
      </div>
    </div>

    <aside class="queue-drawer" class:open={queueOpen} aria-hidden={!queueOpen}>
      <div class="drawer-head">
        <div>
          <h2>{$tr('tv.queueDrawer')}</h2>
          <p class="tiny dim">
            {snap.room.manualOrder ? $tr('tv.orderLocked') : $tr('tv.autoOrder')}
          </p>
        </div>
        <button class="tool-btn" type="button" on:click={() => (queueOpen = false)} aria-label={$tr('tv.closeQueue')}>✕</button>
      </div>

      <div class="drawer-actions">
        <button class="btn btn-secondary btn-sm" type="button" on:click={toggleLock}>
          {$tr(snap.room.queueLocked ? 'room.unlock' : 'room.lock')}
        </button>
        <button class="btn btn-ghost btn-sm" type="button" on:click={rotatePin}>{$tr('tv.rotatePin')}</button>
      </div>
      {#if queue.length > 0}
        <p class="tiny dim drag-hint">{$tr('tv.dragHint')}</p>
      {/if}
      <ol class="drawer-list">
        {#each queue as item, index (item.id)}
          <li
            class="drawer-item"
            draggable="true"
            on:dragstart={(e) => onDragStart(e, item.id)}
            on:dragover|preventDefault
            on:drop={(e) => onDrop(e, item.id)}
            class:dragging={dragId === item.id}
          >
            <span class="drag-handle" aria-hidden="true">⠿</span>
            <span class="drawer-pos mono">{index + 1}</span>
            <img src={item.thumbnailUrl} alt="" />
            <div class="grow">
              <p class="clamp-2 drawer-title">{item.title}</p>
              <p class="tiny dim">{item.nickname} · {item.score} ↑</p>
            </div>
            <button class="remove-btn" type="button" on:click={() => removeSong(item.id)} aria-label={$tr('tv.removeSong')}>✕</button>
          </li>
        {:else}
          <li class="drawer-empty">{$tr('tv.queueEmpty')}</li>
        {/each}
      </ol>

      <div class="drawer-footer">
        {#if !confirmClear}
          <button class="btn btn-ghost btn-sm" type="button" on:click={() => (confirmClear = true)} disabled={queue.length === 0}>
            {$tr('tv.clearQueue')}
          </button>
        {:else}
          <div class="confirm-row">
            <span class="tiny">{$tr('tv.clearConfirm')}</span>
            <button class="btn btn-danger btn-sm" type="button" on:click={clearQueue}>{$tr('common.confirm')}</button>
            <button class="btn btn-ghost btn-sm" type="button" on:click={() => (confirmClear = false)}>{$tr('common.cancel')}</button>
          </div>
        {/if}
        <button class="btn btn-danger btn-sm end-btn" type="button" on:click={armEnd}>
          {confirmEnd ? $tr('tv.endRoomConfirm') : $tr('tv.endRoom')}
        </button>
      </div>
    </aside>

    {#if queueOpen}
      <button class="drawer-scrim" type="button" on:click={() => (queueOpen = false)} aria-label={$tr('tv.closeQueue')}></button>
    {/if}

    {#if toast}
      <div class="toast tv-toast" role="status" aria-live="polite">{toast}</div>
    {/if}
  </div>
{/if}

<style>
  .center-page {
    min-height: 100dvh;
    display: grid;
    place-content: center;
    justify-items: center;
    gap: 1rem;
    padding: 2rem;
    text-align: center;
  }

  .end-card {
    padding: 2rem;
    display: grid;
    gap: 0.9rem;
    justify-items: center;
    max-width: 480px;
  }

  .end-card h1 {
    font-size: 2rem;
  }

  .end-card .btn {
    width: 100%;
  }

  .spinner.large {
    width: 32px;
    height: 32px;
    border-width: 3px;
    color: var(--magenta);
  }

  .tv {
    position: relative;
    height: 100dvh;
    overflow: hidden;
    background: #000;
  }

  .stage {
    position: absolute;
    inset: 0;
    background: radial-gradient(circle at 50% 30%, #24103b, #030107 70%);
  }

  .stage-placeholder {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    background:
      radial-gradient(circle at 20% 20%, rgb(255 45 149 / 0.2), transparent 45%),
      radial-gradient(circle at 80% 80%, rgb(34 211 238 / 0.16), transparent 45%),
      #08050f;
  }

  .stage-orb {
    display: grid;
    place-items: center;
    width: 180px;
    height: 180px;
    border-radius: 50%;
    border: 1px solid var(--border-bright);
    color: var(--magenta);
    font-size: 4rem;
    box-shadow: 0 0 120px rgb(255 45 149 / 0.2);
    animation: pulse 3s infinite;
  }

  .sound-prompt {
    position: absolute;
    z-index: 16;
    left: 50%;
    bottom: 26%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem 1.3rem;
    border-radius: 999px;
    border: 1px solid var(--border-bright);
    background: rgb(8 5 15 / 0.82);
    backdrop-filter: blur(8px);
    color: var(--text);
    cursor: pointer;
    box-shadow: 0 18px 50px rgb(0 0 0 / 0.55), var(--shadow-neon);
    animation: sound-pulse 2.4s ease-in-out infinite;
  }

  .sound-prompt-icon {
    display: grid;
    place-items: center;
    width: 42px;
    height: 42px;
    border-radius: 50%;
    border: 1px solid var(--magenta);
    color: var(--magenta);
    flex: none;
  }

  .sound-prompt-copy {
    display: grid;
    text-align: left;
    gap: 0.1rem;
  }

  .sound-prompt-copy strong {
    font-size: 1rem;
  }

  .sound-prompt-copy small {
    color: var(--muted);
    font-size: 0.78rem;
  }

  @keyframes sound-pulse {
    50% {
      box-shadow: 0 18px 50px rgb(0 0 0 / 0.55), 0 0 34px rgb(255 45 149 / 0.45);
    }
  }

  .vignette {
    position: absolute;
    inset: 0;
    pointer-events: none;
    background:
      linear-gradient(to bottom, rgb(3 1 8 / 0.75), transparent 22%, transparent 65%, rgb(3 1 8 / 0.9)),
      linear-gradient(to right, rgb(3 1 8 / 0.35), transparent 20%, transparent 80%, rgb(3 1 8 / 0.35));
  }

  .tv-top {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    z-index: 10;
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1.5rem;
    padding: 1.4rem 1.7rem;
    pointer-events: none;
  }

  .tv-top > * {
    pointer-events: auto;
  }

  .room-mark {
    display: grid;
    gap: 0.25rem;
    min-width: 0;
    text-shadow: 0 2px 20px rgb(0 0 0 / 0.7);
  }

  .tv-code {
    font-family: var(--mono);
    font-size: clamp(1.7rem, 2.8vw, 2.7rem);
    letter-spacing: 0.12em;
    line-height: 1;
    color: #fff;
  }

  .tv-stat {
    font-size: 0.78rem;
    color: var(--muted);
    font-weight: 700;
  }

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

  .tv-tools {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .tool-btn {
    position: relative;
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: rgb(8 5 15 / 0.65);
    color: var(--text);
    text-decoration: none;
  }

  .tool-btn:hover {
    border-color: var(--cyan);
  }

  .stage-overlay {
    position: absolute;
    inset: 0;
    z-index: 8;
    display: grid;
    place-items: center;
    padding: 2rem;
    pointer-events: none;
  }

  .stage-overlay > * {
    pointer-events: auto;
  }

  .start-card,
  .idle-card,
  .countdown-card {
    display: grid;
    justify-items: center;
    gap: 0.7rem;
    text-align: center;
    padding: 2.2rem;
    border-radius: 24px;
    border: 1px solid var(--border-bright);
    background: rgb(12 7 22 / 0.82);
    backdrop-filter: blur(14px);
    box-shadow: 0 24px 100px rgb(0 0 0 / 0.55);
    max-width: 520px;
  }

  .start-icon,
  .idle-note {
    font-size: 2.4rem;
    line-height: 1;
    margin-bottom: 0.2rem;
  }

  .start-card h2,
  .idle-card h2,
  .countdown-card h2 {
    font-size: clamp(1.5rem, 2.6vw, 2.2rem);
  }

  .start-card p,
  .idle-card p {
    color: var(--muted);
    line-height: 1.55;
    max-width: 38ch;
  }

  .start-card .btn {
    margin-top: 0.6rem;
  }

  .countdown-label {
    font-size: 0.7rem;
    font-weight: 800;
    letter-spacing: 0.2em;
    text-transform: uppercase;
    color: var(--cyan);
  }

  .countdown-singer {
    color: var(--muted);
    font-size: 0.9rem;
  }

  .countdown-number {
    font-family: var(--mono);
    font-size: clamp(4rem, 9vw, 7rem);
    line-height: 0.9;
    color: var(--magenta);
    text-shadow: 0 0 40px rgb(255 45 149 / 0.4);
    margin: 0.2rem 0 0.6rem;
  }

  .skip-banner {
    position: absolute;
    left: 50%;
    bottom: 150px;
    transform: translateX(-50%);
    z-index: 20;
    display: flex;
    align-items: center;
    gap: 1rem;
    width: min(720px, calc(100% - 3rem));
    padding: 0.6rem 0.9rem;
    border-radius: 14px;
    border: 1px solid rgb(255 45 149 / 0.45);
    background: rgb(30 8 24 / 0.92);
    backdrop-filter: blur(12px);
    box-shadow: var(--shadow-neon);
  }

  .skip-copy {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
    font-size: 0.82rem;
  }

  .skip-copy span {
    color: var(--muted);
    font-size: 0.72rem;
  }

  .skip-bar {
    flex: 1;
    height: 6px;
    border-radius: 999px;
    background: rgb(255 255 255 / 0.14);
    overflow: hidden;
  }

  .skip-bar span {
    display: block;
    height: 100%;
    background: linear-gradient(90deg, var(--magenta), var(--cyan));
    transition: width 0.25s ease;
  }

  .skip-time {
    font-size: 0.85rem;
    color: var(--magenta);
  }

  .tv-bottom {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 15;
    display: grid;
    grid-template-columns: minmax(260px, 1.2fr) auto minmax(220px, 1fr);
    align-items: end;
    gap: 1.5rem;
    padding: 1.4rem 1.7rem 1.5rem;
    background: linear-gradient(to top, rgb(5 2 10 / 0.96), rgb(5 2 10 / 0.6) 60%, transparent);
    transition: opacity 0.25s ease, transform 0.25s ease;
  }

  .controls-hidden .tv-bottom {
    opacity: 0;
    transform: translateY(12px);
    pointer-events: none;
  }

  .now-summary {
    display: flex;
    gap: 0.8rem;
    align-items: center;
    min-width: 0;
  }

  .now-thumb {
    width: 116px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 10px;
    border: 1px solid var(--border);
    flex: none;
    background: var(--surface-3);
  }

  .now-meta {
    min-width: 0;
    flex: 1;
  }

  .now-title {
    font-size: 1.08rem;
    font-weight: 800;
    line-height: 1.25;
    margin-bottom: 0.2rem;
  }

  .timeline {
    height: 4px;
    border-radius: 999px;
    background: rgb(255 255 255 / 0.14);
    overflow: hidden;
    margin-top: 0.4rem;
  }

  .timeline span {
    display: block;
    height: 100%;
    background: linear-gradient(90deg, var(--magenta), var(--cyan));
  }

  .playback-controls {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.4rem;
    border-radius: 999px;
    background: rgb(18 10 31 / 0.8);
    border: 1px solid var(--border);
    backdrop-filter: blur(10px);
  }

  .ctrl {
    position: relative;
    display: grid;
    place-items: center;
    width: 48px;
    height: 48px;
    padding: 0;
    border: 0;
    border-radius: 50%;
    background: transparent;
    color: var(--text);
  }

  .ctrl:hover {
    background: rgb(255 255 255 / 0.07);
  }

  .ctrl.primary {
    width: 58px;
    height: 58px;
    background: linear-gradient(135deg, var(--magenta), #ff63b4);
    color: #1a0010;
    box-shadow: var(--shadow-neon);
  }

  .queue-badge {
    position: absolute;
    top: -3px;
    right: -4px;
    min-width: 18px;
    height: 18px;
    padding: 0 4px;
    display: grid;
    place-items: center;
    border-radius: 999px;
    background: var(--cyan);
    color: #001014;
    font-size: 0.62rem;
    font-weight: 900;
  }

  .next-summary {
    text-align: right;
    min-width: 0;
  }

  .next-row {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    margin-top: 0.35rem;
  }

  .next-row img {
    width: 72px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 7px;
    order: 3;
  }

  .next-title {
    font-size: 0.82rem;
    font-weight: 700;
    line-height: 1.3;
  }

  .queue-drawer {
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    z-index: 55;
    width: min(400px, 92vw);
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
    padding: 1.1rem;
    background: rgb(13 7 23 / 0.98);
    border-left: 1px solid var(--border-bright);
    box-shadow: -20px 0 80px rgb(0 0 0 / 0.5);
    transform: translateX(102%);
    transition: transform 0.23s ease;
  }

  .queue-drawer.open {
    transform: translateX(0);
  }

  .drawer-scrim {
    position: absolute;
    inset: 0;
    z-index: 50;
    border: 0;
    background: rgb(0 0 0 / 0.45);
  }

  .drawer-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.8rem;
  }

  .drawer-head h2 {
    font-size: 1.2rem;
  }

  .drawer-actions {
    display: flex;
    gap: 0.4rem;
  }

  .drag-hint {
    margin: -0.3rem 0;
  }

  .drawer-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.45rem;
    overflow-y: auto;
    flex: 1;
    align-content: start;
  }

  .drawer-item {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.45rem;
    border-radius: 10px;
    border: 1px solid var(--border);
    background: var(--surface);
    cursor: grab;
  }

  .drawer-item.dragging {
    opacity: 0.45;
  }

  .drag-handle {
    color: var(--dim);
    font-size: 0.9rem;
    cursor: grab;
  }

  .drawer-pos {
    color: var(--dim);
    font-size: 0.65rem;
    width: 14px;
  }

  .drawer-item img {
    width: 58px;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 6px;
  }

  .drawer-title {
    font-size: 0.76rem;
    font-weight: 700;
    line-height: 1.3;
  }

  .remove-btn {
    border: 0;
    background: transparent;
    color: var(--danger);
    font-size: 0.8rem;
    width: 26px;
    height: 26px;
    border-radius: 6px;
  }

  .remove-btn:hover {
    background: rgb(255 92 92 / 0.12);
  }

  .drawer-empty {
    color: var(--dim);
    font-size: 0.8rem;
    padding: 1rem;
    text-align: center;
  }

  .drawer-footer {
    border-top: 1px solid var(--border);
    padding-top: 0.8rem;
    display: grid;
    gap: 0.5rem;
  }

  .confirm-row {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    flex-wrap: wrap;
  }

  .confirm-row .tiny {
    flex: 1;
    min-width: 120px;
  }

  .end-btn {
    justify-self: start;
  }

  .tv-toast {
    bottom: 145px;
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
    .tv-bottom {
      grid-template-columns: 1fr auto;
      gap: 0.8rem;
    }
    .next-summary {
      display: none;
    }
  }

  @media (max-width: 600px) {
    .tv-top {
      padding: 0.8rem;
    }
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
    .tv-code {
      font-size: 1.3rem;
    }
    .now-thumb {
      width: 74px;
    }
    .now-title {
      font-size: 0.85rem;
    }
    .tv-bottom {
      padding: 0.7rem;
    }
    .playback-controls {
      gap: 0;
      padding: 0.2rem;
    }
    .ctrl {
      width: 40px;
      height: 40px;
    }
    .ctrl.primary {
      width: 46px;
      height: 46px;
    }
    .skip-banner {
      bottom: 104px;
      gap: 0.5rem;
      width: calc(100% - 1.4rem);
      flex-wrap: wrap;
    }
  }
</style>
