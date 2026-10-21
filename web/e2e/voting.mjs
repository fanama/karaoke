import { chromium } from 'playwright';

const base = process.env.BASE_URL || 'http://localhost:8090';
const browser = await chromium.launch({ channel: 'chrome', headless: true });
const errors = [];
const checks = [];

function check(name, ok, detail = '') {
  checks.push({ name, ok });
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` — ${detail}` : ''}`);
}

async function addSong(page, url, nickname) {
  await page.locator('.tabs button').first().click();
  await page.locator('.paste-card input').fill(url);
  await page.locator('.paste-card button').click();
  await page.waitForSelector('#nickname');
  await page.locator('#nickname').fill(nickname);
  await page.getByRole('button', { name: /Add to queue|Ajouter à la file/i }).click();
  await page.waitForSelector('.now-card, .queue-item', { timeout: 20000 });
}

try {
  const tvContext = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const tv = await tvContext.newPage();
  tv.on('pageerror', (e) => errors.push(`tv: ${e.message}`));
  await tv.goto(base, { waitUntil: 'networkidle' });
  await tv.getByRole('button', { name: /Start a karaoke room|Créer une salle/i }).click();
  await tv.waitForURL(/\/tv\//);
  await tv.waitForSelector('.tv-code');
  const code = (await tv.locator('.tv-code').innerText()).trim().toLowerCase();

  const guests = [];
  for (let i = 0; i < 2; i++) {
    const ctx = await browser.newContext({ viewport: { width: 390, height: 844 }, isMobile: true });
    const page = await ctx.newPage();
    page.on('pageerror', (e) => errors.push(`guest${i}: ${e.message}`));
    await page.goto(`${base}/r/${code}`, { waitUntil: 'networkidle' });
    await page.waitForSelector('.queue-panel, .search-panel');
    guests.push(page);
  }

  await addSong(guests[0], 'https://youtu.be/dQw4w9WgXcQ', 'One');
  await addSong(guests[1], 'https://www.youtube.com/watch?v=9bZkp7q19f0', 'Two');
  await tv.waitForFunction(() => document.querySelector('.queue-badge')?.textContent?.trim() === '1', null, { timeout: 10000 });
  check('two guests queue two songs (first auto-starts)', true);

  await tv.waitForTimeout(6000);

  await guests[0].locator('.tabs button').nth(1).click();
  await guests[1].locator('.tabs button').nth(1).click();
  await guests[1].getByRole('button', { name: /Vote to skip|Voter pour passer/i }).click();
  await guests[1].waitForSelector('.vote-live', { timeout: 5000 });
  const progress = await guests[1].locator('.vote-live').innerText();
  check('skip vote starts with tally', /0 of|0 sur/i.test(progress), progress.replace(/\s+/g, ' ').slice(0, 60));

  await guests[0].getByRole('button', { name: /Vote yes|Voter oui/i }).click();
  await guests[1].getByRole('button', { name: /Vote yes|Voter oui/i }).click();
  await tv.waitForSelector('.countdown-card', { timeout: 8000 });
  check('majority skip advances the TV', true);

  const queueAfter = await tv.locator('.queue-badge').innerText();
  check('one song remains while next is in countdown', queueAfter.trim() === '0' || queueAfter.trim() === '1', `badge=${queueAfter}`);
  check('no page errors', errors.length === 0, errors.slice(0, 3).join(' | '));
  await tv.screenshot({ path: '/tmp/fana-vote-tv.png', fullPage: true });
  await guests[1].screenshot({ path: '/tmp/fana-vote-guest.png', fullPage: true });
} catch (err) {
  check('voting flow completed', false, err.message);
  console.error(err);
} finally {
  await browser.close();
}

const failed = checks.filter((c) => !c.ok);
console.log(`\n${checks.length - failed.length}/${checks.length} checks passed`);
process.exit(failed.length ? 1 : 0);
