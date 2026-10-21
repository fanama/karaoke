import type { Status } from '../domain/models';

export interface PlayerTelemetry {
  entryId: string;
  generation: number;
}

export interface PlayerProgress extends PlayerTelemetry {
  position: number;
  duration: number;
}

export interface PlayerFailure extends PlayerTelemetry {
  code: string;
}

export interface PlayerUpdate extends PlayerTelemetry {
  videoId: string | null;
  status: Status;
  elapsed: number;
}

export interface PlayerEvents {
  ended: (detail: PlayerTelemetry) => void;
  error: (detail: PlayerFailure) => void;
  progress: (detail: PlayerProgress) => void;
  blocked: (blocked: boolean) => void;
  soundblocked: (blocked: boolean) => void;
}

export interface VideoPlayerPort {
  update(input: PlayerUpdate): void;
  setMuted(muted: boolean): void;
  forcePlay(): void;
  destroy(): void;
}

export type PlayerFactory = (container: HTMLElement, events: PlayerEvents) => VideoPlayerPort;
