import { chromium } from 'playwright';

// The TV is often opened before the song starts, and the host starts it from a
// phone, so the TV page itself never receives a click. Browsers refuse sound in
// that case; playback must still start (muted) with one tap to unmute.
const base = process.env.BASE_URL || 'http://localhost:8080';
const browser = await chromium.launch({ channel: 'chrome', headless: true });
const checks = [];
const errors = [];

function check(name, ok, detail = '') {
  checks.push({ name, ok });
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` — ${detail}` : ''}`);
}

try {
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const api = ctx.request;

  const createRes = await api.post(`${base}/api/rooms`, { data: {} });
  const room = await createRes.json();
  if (!createRes.ok()) throw new Error(`create room: ${createRes.status()} ${JSON.stringify(room)}`);

  const tv = await ctx.newPage();
  tv.on('pageerror', (e) => errors.push(`tv: ${e.message}`));
  await tv.goto(room.tvUrl, { waitUntil: 'networkidle' });
  await tv.waitForSelector('.tv-code', { timeout: 15000 });

  // Adding a song starts the room automatically without needing any host click.
  const addRes = await api.post(`${base}/api/rooms/${room.code}/songs`, {
    data: { url: 'https://youtu.be/dQw4w9WgXcQ', nickname: 'Host' }
  });
  if (!addRes.ok()) throw new Error(`add song: ${addRes.status()} ${await addRes.text()}`);

  const prompt = tv.locator('.sound-prompt');
  await prompt.waitFor({ state: 'visible', timeout: 25000 });
  check('muted autoplay chip appears without a gesture', true);
  check('no hard autoplay block', (await tv.locator('.blocked-overlay').count()) === 0);

  const frame = tv.frames().find((f) => f.url().includes('youtube.com/embed'));
  const readVideo = () =>
    frame
      ? frame.evaluate(() => {
          const v = document.querySelector('video');
          return v ? { paused: v.paused, muted: v.muted } : null;
        })
      : null;

  let before = await readVideo();
  for (let i = 0; i < 12 && (before?.paused !== false || before?.muted !== true); i++) {
    await tv.waitForTimeout(250);
    before = await readVideo();
  }
  check(
    'player is playing while muted',
    before?.paused === false && before?.muted === true,
    JSON.stringify(before)
  );
  await tv.screenshot({ path: '/tmp/fana-autoplay-muted.png', fullPage: true });

  await prompt.click();
  await tv.waitForTimeout(600);
  check('chip disappears after the tap', (await tv.locator('.sound-prompt').count()) === 0);
  const after = await readVideo();
  check('tap turns the sound on', after?.muted === false, JSON.stringify(after));
  await tv.screenshot({ path: '/tmp/fana-autoplay-sound.png', fullPage: true });

  check('no page errors', errors.length === 0, errors.slice(0, 3).join(' | '));
  await api.post(`${base}/api/rooms/${room.code}/end`, { data: {} });
} catch (err) {
  check('autoplay flow completed', false, err.message);
  console.error(err);
} finally {
  await browser.close();
}

const failed = checks.filter((c) => !c.ok);
console.log(`\n${checks.length - failed.length}/${checks.length} checks passed`);
process.exit(failed.length ? 1 : 0);
