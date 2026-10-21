<script lang="ts">
  import { createEventDispatcher, onDestroy, onMount } from 'svelte';
  import type { Status } from '../../lib/types';
  import { tr } from '../../lib/i18n';

  export let videoId: string | null = null;
  export let entryId = '';
  export let generation = 0;
  export let status: Status = 'idle';
  export let elapsed = 0;
  export let muted = false;

  const dispatch = createEventDispatcher<{
    ended: { entryId: string; generation: number };
    error: { entryId: string; generation: number; code: string };
    progress: { entryId: string; generation: number; position: number; duration: number };
    blocked: boolean;
    soundblocked: boolean;
  }>();

  let container: HTMLDivElement;
  let player: any = null;
  let apiReady = false;
  let apiLoading = false;
  let loadedVideo = '';
  let ready = false;
  let lastStatus: Status | '' = '';
  let blocked = false;
  // Set when the browser refused to start with sound and we fell back to muted
  // playback, which every browser allows without a user gesture.
  let mutedByAutoplay = false;
  let mutedForPlayer = false;
  // The user explicitly asked for audio; never mute them again afterwards.
  let soundUnlocked = false;
  let soundRecoveries = 0;
  let progressTimer: ReturnType<typeof setInterval> | null = null;
  let blockedTimer: ReturnType<typeof setTimeout> | null = null;
  let destroyed = false;

  function loadAPI(): Promise<void> {
    return new Promise((resolve) => {
      const w = window as any;
      if (w.YT?.Player) {
        resolve();
        return;
      }
      w.__fanaYouTubeReady = () => {
        apiReady = true;
        resolve();
      };
      if (apiLoading) {
        const check = setInterval(() => {
          if (w.YT?.Player) {
            clearInterval(check);
            apiReady = true;
            resolve();
          }
        }, 150);
        return;
      }
      apiLoading = true;
      w.onYouTubeIframeAPIReady = () => {
        apiReady = true;
        w.__fanaYouTubeReady?.();
        resolve();
      };
      const script = document.createElement('script');
      script.src = 'https://www.youtube.com/iframe_api';
      script.async = true;
      script.onerror = () => {
        apiReady = false;
        dispatch('error', { entryId, generation, code: 'api' });
      };
      document.head.appendChild(script);
    });
  }

  function createPlayer() {
    if (!apiReady || !videoId || player || destroyed) return;
    ready = false;
    player = new (window as any).YT.Player(container, {
      videoId,
      playerVars: {
        autoplay: 1,
        controls: 0,
        disablekb: 1,
        fs: 1,
        iv_load_policy: 3,
        modestbranding: 1,
        playsinline: 1,
        rel: 0,
        origin: location.origin
      },
      events: {
        onReady: () => {
          ready = true;
          loadedVideo = videoId ?? '';
          applyState(true);
          startProgress();
        },
        onStateChange: (event: any) => {
          const state = event.data;
          if (state === 0) {
            dispatch('ended', { entryId, generation });
          } else if (state === 1) {
            blocked = false;
            dispatch('blocked', false);
          } else if (state === 2) {
            // Paused by YouTube UI; the server remains authoritative.
          }
        },
        onError: (event: any) => {
          dispatch('error', {
            entryId,
            generation,
            code: event.data === 101 || event.data === 150 ? 'embed' : 'unavailable'
          });
        }
      }
    });
  }

  // status and elapsed are passed in so the reactive call below re-runs when the
  // server changes playback state, not only when the video changes.
  function applyState(videoChanged = false, statusNow: Status = status, elapsedNow = elapsed) {
    if (!ready || !player || !videoId) return;
    // getVideoData() can briefly return an empty id while the iframe loads;
    // fall back to what we last loaded so a transient miss cannot re-cue and
    // pause a song that is already playing.
    const current = player.getVideoData?.()?.video_id || loadedVideo;
    if (current !== videoId) {
      loadedVideo = videoId;
      lastStatus = '';
      soundRecoveries = 0;
      blocked = false;
      if (statusNow === 'playing') {
        player.loadVideoById({ videoId, startSeconds: Math.max(0, elapsedNow) });
        player.playVideo?.();
        scheduleBlockedCheck();
        return;
      }
      player.cueVideoById({ videoId, startSeconds: Math.max(0, elapsedNow) });
      if (statusNow !== 'paused') player.pauseVideo();
    }
    if (statusNow !== lastStatus || videoChanged) {
      if (statusNow === 'playing') {
        const pos = player.getCurrentTime?.() ?? 0;
        if (Math.abs(pos - elapsedNow) > 2) {
          player.seekTo(Math.max(0, elapsedNow), true);
        }
        player.playVideo();
        scheduleBlockedCheck();
      } else if (statusNow === 'paused' || statusNow === 'countdown') {
        player.pauseVideo();
        if (statusNow === 'paused' && Math.abs((player.getCurrentTime?.() ?? 0) - elapsedNow) > 2) {
          player.seekTo(Math.max(0, elapsedNow), true);
        }
      }
      lastStatus = statusNow;
    }
  }

  function applyMute() {
    if (!player) return;
    if (muted || mutedByAutoplay) player.mute?.();
    else player.unMute?.();
  }

  const STEP = 750;

  function scheduleBlockedCheck(attempt = 0) {
    if (blockedTimer) clearTimeout(blockedTimer);
    blockedTimer = setTimeout(() => {
      if (destroyed || !player || status !== 'playing') return;
      const state = player.getPlayerState?.();
      const position = player.getCurrentTime?.() ?? 0;
      // Position is the only trustworthy signal: a browser that refused to
      // start reports state 3 (buffering) with the clock frozen at zero.
      if (state === 1 || position > 0.5) return;
      const waited = (attempt + 1) * STEP;
      if (!mutedByAutoplay && !soundUnlocked && waited >= 2 * STEP && soundRecoveries < 3) {
        // Sound was refused. Muted autoplay is allowed, so keep the song
        // rolling and ask for a single tap to bring the audio back.
        soundRecoveries += 1;
        mutedByAutoplay = true;
        applyMute();
        player.playVideo?.();
        setTimeout(() => {
          if (!destroyed && status === 'playing') {
            applyMute();
            player?.playVideo?.();
          }
        }, 250);
        dispatch('soundblocked', true);
        scheduleBlockedCheck(attempt + 1);
        return;
      }
      if (waited < 25 * STEP) {
        player.playVideo?.();
        scheduleBlockedCheck(attempt + 1);
        return;
      }
      blocked = true;
      dispatch('blocked', true);
    }, STEP);
  }

  function enableSound() {
    soundUnlocked = true;
    mutedByAutoplay = false;
    blocked = false;
    dispatch('blocked', false);
    dispatch('soundblocked', false);
    player?.setVolume?.(100);
    player?.unMute?.();
    if (status === 'playing') {
      player?.playVideo?.();
      scheduleBlockedCheck();
    }
  }

  function onGesture() {
    // Any tap or key press counts as the activation browsers ask for.
    if (mutedByAutoplay) enableSound();
  }

  function startProgress() {
    if (progressTimer) return;
    progressTimer = setInterval(() => {
      if (!player || status !== 'playing') return;
      const position = player.getCurrentTime?.() ?? 0;
      const duration = player.getDuration?.() ?? 0;
      if (position > 0) {
        dispatch('progress', { entryId, generation, position, duration });
      }
    }, 5000);
  }

  export function forcePlay() {
    enableSound();
    player?.playVideo?.();
    scheduleBlockedCheck();
    startProgress();
  }

  export function setMuted(value: boolean) {
    if (value) {
      player?.mute?.();
      return;
    }
    soundUnlocked = true;
    mutedByAutoplay = false;
    dispatch('soundblocked', false);
    player?.setVolume?.(100);
    player?.unMute?.();
  }

  onMount(() => {
    window.addEventListener('pointerdown', onGesture, true);
    window.addEventListener('keydown', onGesture, true);
    loadAPI().then(() => {
      createPlayer();
      startProgress();
    });
  });

  onDestroy(() => {
    destroyed = true;
    window.removeEventListener('pointerdown', onGesture, true);
    window.removeEventListener('keydown', onGesture, true);
    if (progressTimer) clearInterval(progressTimer);
    if (blockedTimer) clearTimeout(blockedTimer);
    try {
      player?.destroy?.();
    } catch {
      // iframe already gone
    }
  });
  $: if (apiReady && videoId && !player) createPlayer();
  $: if (apiReady && player && videoId) applyState(false, status, elapsed);
  $: mutedForPlayer = muted || mutedByAutoplay;
  $: if (player) {
    if (mutedForPlayer) player.mute?.();
    else player.unMute?.();
  }
</script>

<div class="player-wrap">
  <div class="player" bind:this={container}></div>
  {#if blocked}
    <button class="blocked-overlay" type="button" on:click={forcePlay}>
      <span class="tap-ring" aria-hidden="true">▶</span>
      <strong>{$tr('tv.tapResume')}</strong>
      <small>{$tr('tv.tapResumeBody')}</small>
    </button>
  {/if}
</div>

<style>
  .player-wrap {
    position: absolute;
    inset: 0;
    background: #000;
  }

  .player,
  .player :global(iframe) {
    width: 100%;
    height: 100%;
    border: 0;
  }

  .player {
    pointer-events: auto;
  }

  .blocked-overlay {
    position: absolute;
    inset: 0;
    z-index: 5;
    display: grid;
    place-content: center;
    justify-items: center;
    gap: 0.5rem;
    border: 0;
    background: rgb(8 5 15 / 0.72);
    backdrop-filter: blur(5px);
    color: var(--text);
    text-align: center;
    padding: 1rem;
  }

  .tap-ring {
    display: grid;
    place-items: center;
    width: 84px;
    height: 84px;
    border-radius: 50%;
    border: 2px solid var(--magenta);
    color: var(--magenta);
    font-size: 1.8rem;
    box-shadow: var(--shadow-neon);
    margin-bottom: 0.5rem;
  }

  .blocked-overlay strong {
    font-size: clamp(1.2rem, 2.4vw, 1.8rem);
  }

  .blocked-overlay small {
    color: var(--muted);
    font-size: 0.9rem;
  }
</style>
