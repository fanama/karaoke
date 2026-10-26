import { chromium } from 'playwright';

const base = process.env.BASE_URL || 'http://localhost:8090';
const browser = await chromium.launch({ channel: 'chrome', headless: true });
const errors = [];
const checks = [];

function check(name, ok, detail = '') {
  checks.push({ name, ok });
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` — ${detail}` : ''}`);
}

try {
  const desktop = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const welcome = await desktop.newPage();
  welcome.on('pageerror', (e) => errors.push(`welcome: ${e.message}`));
  await welcome.goto(base, { waitUntil: 'networkidle' });
  const overflow = await welcome.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth
  );
  check('welcome has no horizontal overflow', overflow <= 0, `${overflow}px`);

  await welcome.getByRole('button', { name: /Start a karaoke room|Créer une salle/i }).click();
  await welcome.waitForURL(/\/tv\//);
  await welcome.waitForSelector('.tv-code');
  const code = (await welcome.locator('.tv-code').innerText()).trim().toLowerCase();
  const pin = (await welcome.locator('.pin').innerText()).trim();

  const layout = await welcome.evaluate(() => {
    const top = document.querySelector('.tv-top')?.getBoundingClientRect();
    const bottom = document.querySelector('.tv-bottom')?.getBoundingClientRect();
    const join = document.querySelector('.join-panel')?.getBoundingClientRect();
    return {
      topBottom: top && bottom ? Math.round(bottom.top - top.bottom) : -1,
      joinRight: join?.right,
      viewport: window.innerWidth
    };
  });
  check('TV top and bottom do not overlap', layout.topBottom > 250, `gap=${layout.topBottom}px`);
  check('TV join panel stays on screen', layout.joinRight <= layout.viewport, `right=${layout.joinRight}`);

  const mobile = await browser.newContext({
    viewport: { width: 390, height: 844 },
    isMobile: true,
    hasTouch: true,
    locale: 'fr-FR'
  });
  const guest = await mobile.newPage();
  guest.on('pageerror', (e) => errors.push(`guest: ${e.message}`));
  await guest.goto(`${base}/r/${code}`, { waitUntil: 'networkidle' });
  await guest.waitForSelector('.room-code');
  const guestOverflow = await guest.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth
  );
  check('guest phone has no horizontal overflow', guestOverflow <= 0, `${guestOverflow}px`);
  const french = await guest.locator('.tabs button').first().innerText();
  check('French is selected from browser locale', /rechercher/i.test(french), french.trim());

  const search = guest.locator('.search-input');
  await search.fill('bohemian rhapsody');
  await search.press('Enter');
  let searchOk = true;
  try {
    await guest.waitForSelector('.result', { timeout: 25000 });
  } catch {
    searchOk = false;
  }
  const count = await guest.locator('.result').count();
  const stale = await guest.locator('.notice-strip.cyan').count();
  check('YouTube search returns karaoke results', searchOk && count > 0, `results=${count} stale=${stale}`);
  if (count > 0) {
    const first = (await guest.locator('.result-title').first().innerText()).toLowerCase();
    check('search auto-prefers karaoke videos', first.includes('karaoke') || first.includes('instrumental'), first.slice(0, 60));
  }
  await guest.screenshot({ path: '/tmp/fana-audit-search-fr.png', fullPage: true });

  await guest.locator('.lang-toggle button').first().click();
  const english = await guest.locator('.tabs button').first().innerText();
  check('language toggle switches to English', /search/i.test(english), english.trim());

  await welcome.waitForTimeout(300);
  check('no page errors', errors.length === 0, errors.slice(0, 3).join(' | '));
} catch (err) {
  check('audit completed', false, err.message);
  console.error(err);
} finally {
  await browser.close();
}

const failed = checks.filter((c) => !c.ok);
console.log(`\n${checks.length - failed.length}/${checks.length} checks passed`);
process.exit(failed.length ? 1 : 0);
