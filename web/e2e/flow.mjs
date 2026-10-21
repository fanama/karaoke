import { chromium } from 'playwright';

const base = process.env.BASE_URL || 'http://localhost:8090';
const browser = await chromium.launch({ channel: 'chrome', headless: true });
const errors = [];
const results = [];

function note(name, ok, detail = '') {
  results.push({ name, ok, detail });
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` — ${detail}` : ''}`);
}

try {
  const tvContext = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const tv = await tvContext.newPage();
  tv.on('pageerror', (err) => errors.push(`TV pageerror: ${err.message}`));
  tv.on('console', (msg) => {
    if (msg.type() === 'error' && !msg.text().includes('youtube')) {
      errors.push(`TV console: ${msg.text()}`);
    }
  });

  await tv.goto(base, { waitUntil: 'networkidle' });
  note('welcome renders', (await tv.locator('h1').count()) > 0);
  await tv.screenshot({ path: '/tmp/fana-welcome.png', fullPage: true });

  await tv.getByRole('button', { name: /Start a karaoke room|Créer une salle/i }).click();
  await tv.waitForURL(/\/tv\//, { timeout: 15000 });
  await tv.waitForSelector('.tv-code', { timeout: 15000 });
  const code = (await tv.locator('.tv-code').innerText()).trim();
  const pin = (await tv.locator('.pin').innerText()).trim();
  note('room created with code and pin', code.length >= 5 && /^\d{4}$/.test(pin), `${code} / ${pin}`);

  const qrStatus = await tv.locator('.qr-box img').evaluate((img) => img.complete && img.naturalWidth > 0);
  note('QR image loads', qrStatus);
  await tv.screenshot({ path: '/tmp/fana-tv-idle.png', fullPage: true });

  const guestContext = await browser.newContext({
    viewport: { width: 390, height: 844 },
    isMobile: true,
    hasTouch: true
  });
  const guest = await guestContext.newPage();
  guest.on('pageerror', (err) => errors.push(`Guest pageerror: ${err.message}`));
  guest.on('console', (msg) => {
    if (msg.type() === 'error' && !msg.text().includes('youtube')) {
      errors.push(`Guest console: ${msg.text()}`);
    }
  });

  await guest.goto(`${base}/r/${code.toLowerCase()}`, { waitUntil: 'networkidle' });
  await guest.waitForSelector('.room-code', { timeout: 15000 });
  const guestCode = (await guest.locator('.room-code').innerText()).trim();
  note('guest joins room', guestCode === code);

  const compact = await guest.request.get(
    `${base}/api/rooms/resolve?code=${encodeURIComponent(code.replace(/-/g, ''))}`
  );
  const compactBody = await compact.json();
  note('compact code resolves', compact.ok() && compactBody.code === code.toLowerCase());
  await guest.screenshot({ path: '/tmp/fana-guest-search.png', fullPage: true });

  await guest.locator('.paste-card input').fill('https://youtu.be/dQw4w9WgXcQ');
  await guest.locator('.paste-card button').click();
  await guest.waitForSelector('#nickname', { timeout: 5000 });
  await guest.locator('#nickname').fill('Playwright');
  await guest.getByRole('button', { name: /Add to queue|Ajouter à la file/i }).click();
  await guest.waitForSelector('.now-card', { timeout: 20000 });
  const nowTitle = await guest.locator('.now-card h2').first().innerText();
  note('first song auto-starts on TV', nowTitle.length > 3, nowTitle.slice(0, 50));
  await guest.screenshot({ path: '/tmp/fana-guest-queue.png', fullPage: true });

  await guest.locator('.lock-btn').click();
  await guest.waitForSelector('#pin', { timeout: 5000 });
  await guest.locator('#pin').fill(pin);
  await guest.getByRole('button', { name: /Unlock|Déverrouiller/i }).last().click();
  await guest.waitForSelector('.host-bar', { timeout: 8000 });
  note('host unlock shows host bar', true);

  // Add a second song so the TV has a next-up entry and a non-empty drawer.
  await guest.locator('.tabs button').first().click();
  await guest.locator('.paste-card input').fill('https://www.youtube.com/watch?v=9bZkp7q19f0');
  await guest.locator('.paste-card button').click();
  await guest.waitForSelector('#nickname', { timeout: 5000 });
  await guest.locator('#nickname').fill('Second');
  await guest.getByRole('button', { name: /Add to queue|Ajouter à la file/i }).click();
  await guest.waitForSelector('.queue-item', { timeout: 20000 });
  note('second song queued', true);
  await guest.screenshot({ path: '/tmp/fana-guest-host.png', fullPage: true });

  await tv.waitForFunction(() => document.querySelectorAll('.queue-badge').length > 0, null, {
    timeout: 10000
  });
  const badge = await tv.locator('.queue-badge').innerText();
  note('TV receives live queue', badge.trim() === '1', `badge=${badge}`);

  await tv.waitForFunction(() => document.querySelectorAll('iframe').length > 0, null, {
    timeout: 15000
  });
  const iframeCount = await tv.locator('iframe').count();
  note('YouTube iframe mounted automatically', iframeCount > 0, `iframes=${iframeCount}`);
  await tv.screenshot({ path: '/tmp/fana-tv-playing.png', fullPage: true });

  // Playback must start on its own: either with sound (the page already had a
  // gesture) or muted, in which case a tap-to-unmute chip replaces the old
  // hard "tap to resume" wall.
  await tv.waitForTimeout(6000);
  const ytFrame = tv.frames().find((f) => f.url().includes('youtube.com/embed'));
  const ytVideo = ytFrame
    ? await ytFrame.evaluate(() => {
        const v = document.querySelector('video');
        return v ? { paused: v.paused, muted: v.muted } : null;
      })
    : null;
  note(
    'YouTube player autoplays (not awaiting a gesture)',
    ytVideo?.paused === false,
    `video=${JSON.stringify(ytVideo)}`
  );
  const hardBlocked = await tv.locator('.blocked-overlay').count();
  note('no hard autoplay block', hardBlocked === 0, `overlays=${hardBlocked}`);
  const prompt = tv.locator('.sound-prompt');
  if (await prompt.count()) {
    await prompt.click();
    await tv.waitForTimeout(500);
    note('tap hides the sound prompt', (await tv.locator('.sound-prompt').count()) === 0);
    const afterTap = ytFrame
      ? await ytFrame.evaluate(() => {
          const v = document.querySelector('video');
          return v ? { paused: v.paused, muted: v.muted } : null;
        })
      : null;
    note('tap unmutes the video', afterTap?.muted === false, `video=${JSON.stringify(afterTap)}`);
  } else {
    note('autoplay kept the sound on', true);
  }
  await tv.screenshot({ path: '/tmp/fana-tv-autoplay.png', fullPage: true });

  // Keyboard shortcut: proves the queue stays reachable during playback.
  await tv.keyboard.press('q');
  await tv.waitForTimeout(300);
  const drawerItems = await tv.locator('.drawer-item').count();
  note('TV queue drawer opens', drawerItems >= 1, `items=${drawerItems}`);
  await tv.screenshot({ path: '/tmp/fana-tv-queue.png', fullPage: true });

  const endBtn = tv.locator('.end-btn');
  await endBtn.click();
  await tv.waitForTimeout(150);
  await endBtn.click();
  await guest.waitForSelector('.end-card', { timeout: 10000 });
  note('ending room closes guest view', true);

  note('no page errors', errors.length === 0, errors.slice(0, 4).join(' | '));
} catch (err) {
  note('flow completed', false, err.message);
  console.error(err);
} finally {
  await browser.close();
}

const failed = results.filter((r) => !r.ok);
console.log(`\n${results.length - failed.length}/${results.length} checks passed`);
process.exit(failed.length ? 1 : 0);
