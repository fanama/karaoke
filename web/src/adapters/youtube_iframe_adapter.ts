import {
  BLOCKED_CHECK_MS,
  DRIFT_TOLERANCE_SECONDS,
  PROGRESS_INTERVAL_MS,
  autoplayDecision
} from '../domain/player_policy';
import type {
  PlayerEvents,
  PlayerUpdate,
  VideoPlayerPort
} from '../ports/player_port';

export function createYouTubePlayer(container: HTMLElement, events: PlayerEvents): VideoPlayerPort {
  const player = new YouTubePlayer(container, events);
  player.mount();
  return player;
}

class YouTubePlayer implements VideoPlayerPort {
  private yt: any = null;
  private ready = false;
  private destroyed = false;
  private loadedVideo = '';
  private lastStatus = '';
  private muted = false;
  private mutedByAutoplay = false;
  private soundUnlocked = false;
  private recoveries = 0;
  private progressTimer: ReturnType<typeof setInterval> | null = null;
  private blockedTimer: ReturnType<typeof setTimeout> | null = null;
  private current: PlayerUpdate = {
    videoId: null,
    entryId: '',
    generation: 0,
    status: 'idle',
    elapsed: 0
  };

  constructor(
    private readonly container: HTMLElement,
    private readonly events: PlayerEvents
  ) {}

  mount(): void {
    this.loadAPI().then(() => {
      this.ensurePlayer();
      this.startProgress();
    });
  }

  destroy(): void {
    this.destroyed = true;
    if (this.progressTimer) clearInterval(this.progressTimer);
    if (this.blockedTimer) clearTimeout(this.blockedTimer);
    try {
      this.yt?.destroy?.();
    } catch {
      // iframe already gone
    }
  }

  update(input: PlayerUpdate): void {
    this.current = input;
    if (!this.ready || !this.yt || !input.videoId) return;
    if (this.loadedVideo !== input.videoId) this.changeVideo(input);
    else this.syncStatus(input);
  }

  setMuted(value: boolean): void {
    this.muted = value;
    if (!value) {
      this.soundUnlocked = true;
      this.mutedByAutoplay = false;
      this.events.soundblocked(false);
      this.yt?.setVolume?.(100);
    }
    this.applyMute();
  }

  forcePlay(): void {
    this.enableSound();
    this.yt?.playVideo?.();
    this.scheduleBlockedCheck(0);
    this.startProgress();
  }

  private loadAPI(): Promise<void> {
    return new Promise((resolve) => {
      const w = window as any;
      if (w.YT?.Player) {
        resolve();
        return;
      }
      w.onYouTubeIframeAPIReady = () => resolve();
      if (!document.querySelector('script[data-fana-yt]')) this.injectScript();
    });
  }

  private injectScript(): void {
    const script = document.createElement('script');
    script.src = 'https://www.youtube.com/iframe_api';
    script.async = true;
    script.dataset.fanaYt = 'true';
    script.onerror = () => this.events.error({ ...this.telemetry(), code: 'api' });
    document.head.appendChild(script);
  }

  private ensurePlayer(): void {
    if (this.yt || this.destroyed || !this.current.videoId) return;
    this.yt = new (window as any).YT.Player(this.container, {
      videoId: this.current.videoId,
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
        onReady: () => this.onReady(),
        onStateChange: (event: any) => this.onStateChange(event),
        onError: (event: any) => this.onError(event)
      }
    });
  }

  private onReady(): void {
    this.ready = true;
    this.loadedVideo = this.current.videoId ?? '';
    this.applyMute();
    this.syncStatus(this.current);
    this.startProgress();
  }

  private onStateChange(event: any): void {
    if (event.data === 0) this.events.ended(this.telemetry());
    else if (event.data === 1) this.events.blocked(false);
  }

  private onError(event: any): void {
    const code = event.data === 101 || event.data === 150 ? 'embed' : 'unavailable';
    this.events.error({ ...this.telemetry(), code });
  }

  private telemetry(): { entryId: string; generation: number } {
    return { entryId: this.current.entryId, generation: this.current.generation };
  }

  private changeVideo(input: PlayerUpdate): void {
    this.loadedVideo = input.videoId ?? '';
    this.lastStatus = '';
    this.recoveries = 0;
    this.events.blocked(false);
    const startSeconds = Math.max(0, input.elapsed);
    if (input.status === 'playing') {
      this.yt.loadVideoById({ videoId: input.videoId, startSeconds });
      this.yt.playVideo?.();
      this.scheduleBlockedCheck(0);
      return;
    }
    this.yt.cueVideoById({ videoId: input.videoId, startSeconds });
    if (input.status !== 'paused') this.yt.pauseVideo();
  }

  private syncStatus(input: PlayerUpdate): void {
    const videoChanged = this.loadedVideo !== input.videoId;
    if (input.status === this.lastStatus && !videoChanged) return;
    const position = this.yt.getCurrentTime?.() ?? 0;
    if (input.status === 'playing') {
      if (Math.abs(position - input.elapsed) > DRIFT_TOLERANCE_SECONDS) {
        this.yt.seekTo(Math.max(0, input.elapsed), true);
      }
      this.yt.playVideo();
      this.scheduleBlockedCheck(0);
    } else if (input.status === 'paused' || input.status === 'countdown') {
      this.yt.pauseVideo();
      if (input.status === 'paused' && Math.abs(position - input.elapsed) > DRIFT_TOLERANCE_SECONDS) {
        this.yt.seekTo(Math.max(0, input.elapsed), true);
      }
    }
    this.lastStatus = input.status;
  }

  private applyMute(): void {
    if (this.muted || this.mutedByAutoplay) this.yt?.mute?.();
    else this.yt?.unMute?.();
  }

  private scheduleBlockedCheck(attempt: number): void {
    if (this.blockedTimer) clearTimeout(this.blockedTimer);
    this.blockedTimer = setTimeout(() => this.recheckPlayback(attempt), BLOCKED_CHECK_MS);
  }

  private recheckPlayback(attempt: number): void {
    console.log('recheck', attempt, this.current.status, this.yt?.getPlayerState?.(), this.yt?.getCurrentTime?.());
    if (this.destroyed || !this.yt || this.current.status !== 'playing') return;
    const position = this.yt.getCurrentTime?.() ?? 0;
    const decision = autoplayDecision({
      attempt,
      position,
      mutedByAutoplay: this.mutedByAutoplay,
      soundUnlocked: this.soundUnlocked,
      recoveries: this.recoveries
    });
    console.log('decision', decision);
    if (decision === 'wait') return;
    if (decision === 'mute') return this.fallbackToMuted(attempt);
    if (decision === 'play') {
      this.yt.playVideo?.();
      return this.scheduleBlockedCheck(attempt + 1);
    }
    this.events.blocked(true);
  }

  private fallbackToMuted(attempt: number): void {
    this.recoveries += 1;
    this.mutedByAutoplay = true;
    this.applyMute();
    this.yt.playVideo?.();
    setTimeout(() => {
      if (!this.destroyed && this.current.status === 'playing') {
        this.applyMute();
        this.yt?.playVideo?.();
      }
    }, 250);
    this.events.soundblocked(true);
    this.scheduleBlockedCheck(attempt + 1);
  }

  private startProgress(): void {
    if (this.progressTimer) return;
    this.progressTimer = setInterval(() => {
      if (!this.yt || this.current.status !== 'playing') return;
      const position = this.yt.getCurrentTime?.() ?? 0;
      const duration = this.yt.getDuration?.() ?? 0;
      if (position > 0) this.events.progress({ ...this.telemetry(), position, duration });
    }, PROGRESS_INTERVAL_MS);
  }

  private enableSound(): void {
    this.soundUnlocked = true;
    this.mutedByAutoplay = false;
    this.events.blocked(false);
    this.events.soundblocked(false);
    this.yt?.setVolume?.(100);
    this.yt?.unMute?.();
  }
}
