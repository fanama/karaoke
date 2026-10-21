export interface ClipboardPort {
  writeText(value: string): Promise<void>;
}

export interface FullscreenPort {
  toggle(): void;
}

export interface StoragePort {
  read(key: string): string | null;
  write(key: string, value: string): void;
}

export interface NavigationPort {
  assign(url: string): void;
}

export interface ClockPort {
  now(): number;
}
