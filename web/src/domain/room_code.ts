/** Normalizes what a guest typed into a room code: "BLUE MIC" -> "blue-mic". */
export function normalizeCode(raw: string): string {
  return raw
    .trim()
    .toLowerCase()
    .replace(/[_\s]+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

const VIDEO_ID = /^[A-Za-z0-9_-]{11}$/;

/** Cheap client-side check; the server remains authoritative. */
export function looksLikeYouTubeUrl(raw: string): boolean {
  try {
    const url = new URL(raw.trim());
    const host = url.hostname.toLowerCase();
    if (!['youtube.com', 'www.youtube.com', 'm.youtube.com', 'youtu.be'].includes(host)) return false;
    if (host === 'youtu.be') return VIDEO_ID.test(url.pathname.replace(/^\//, ''));
    return VIDEO_ID.test(url.searchParams.get('v') ?? '');
  } catch {
    return false;
  }
}

export function isValidNickname(value: string): boolean {
  const name = value.trim();
  return name.length >= 1 && name.length <= 24;
}

export function isShortLink(raw: string): boolean {
  return /youtube\.com\/shorts\//i.test(raw);
}
