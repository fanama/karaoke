import type {
  ClipboardPort,
  ClockPort,
  FullscreenPort,
  NavigationPort,
  StoragePort
} from '../ports/browser_port';

export const clipboardAdapter: ClipboardPort = {
  writeText: (value) => navigator.clipboard.writeText(value)
};

export const fullscreenAdapter: FullscreenPort = {
  toggle() {
    if (document.fullscreenElement) {
      document.exitFullscreen?.();
      return;
    }
    document.documentElement.requestFullscreen?.();
  }
};

export const navigationAdapter: NavigationPort = {
  assign: (url) => location.assign(url)
};

export const clockAdapter: ClockPort = {
  now: () => Date.now()
};

export const storageAdapter: StoragePort = {
  read(key) {
    try {
      return localStorage.getItem(key);
    } catch {
      return null;
    }
  },
  write(key, value) {
    try {
      localStorage.setItem(key, value);
    } catch {
      // storage can be unavailable in private mode
    }
  }
};
