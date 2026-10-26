# Remove the AI slop — spec

Status: implemented. See §12 for what shipped, the deviations from this plan, and the
verification results.
Scope owner: Fanamperantsoa Rakotoasimbola
Source of the rule set: the `no-ui-slop` skill (pasted in the request; kept outside the repo, restated in `web/DESIGN.md` in this product's words).

---

## 1. The problem

The UI was built screen by screen and drifted toward the defaults the skill names: decorative
gradients, glow instead of hierarchy, blurred glass panels, four competing accent colors, pills
everywhere, and a row of three identical icon-title-sentence cards on the landing page. None of it
is broken; it is generic. It also reads as one undifferentiated "neon AI app" rather than a
karaoke product, and there is no written rule set, so every new screen re-invents the same look.

### Evidence gathered from the codebase (before any change)

| Pattern | Count | Where |
|---|---|---|
| Gradient declarations (`linear-gradient` etc.) | 24 occurrences / 10 files | heaviest: `ui/tv/TvRoomView.svelte` (5), `styles/components.css` (4), `ui/welcome/WelcomeStagePreview.svelte` (4), `ui/tv/TvBottomBar.svelte` (3) |
| `backdrop-filter` (glass) | 9 files | `components.css` (modal backdrop), `ui/tv/*` (5), `ui/guest/GuestHeader`, `GuestHostBar` |
| `box-shadow` / blur glow | 22 declarations | `components.css` (7), `TvRoomView`, `TvSoundPrompt`, `WelcomeStagePreview` |
| Emoji / typographic glyphs used as icons | 17 occurrences | `♪ ✕ 🎤 ⌕ ◎ ⤨ 🔒` across welcome, guest, tv, admin |
| `border-radius: 999px` or `50%` | 31 occurrences | 14 pills, 17 circles — including `.card`, `.btn`, `.badge` |
| Distinct `border-radius` values | 14 | 1, 4, 6, 7, 8, 9, 10, 12, 13, 14, 16, 22, 24 px + pill/percent |
| Distinct `font-size` values | 28 | 0.62 → 4rem, many within 1px of each other (0.78/0.8/0.82, 0.85/0.86/0.88, 0.9/0.92) |
| Distinct `gap` values | 11+ | 0.35, 0.4, 0.45, 0.5, 0.55, 0.6, 0.65, 0.7, 0.8, 0.85, 1rem |
| Hardcoded hex in `web/src/ui` | 23 | 9 components |
| Accent colors in tokens | 4 | magenta, cyan, violet, lime |

Current token file: `web/src/styles/tokens.css` (`--bg #08050f`, `--shadow-neon`, `--shadow-cyan`,
`--radius 18px`, `color-scheme: dark`). Shared components: `web/src/styles/components.css`.
No webfonts are loaded (`@font-face` / Google Fonts: zero hits), so the app ships with system
stacks only — which the redesign keeps.

### What "slop" is not

Not every criticism applies. The existing copy is mostly concrete ("Ask the host for the code on
the TV"), the flows are real, the empty states exist, there are no fake stats or testimonials, and
the e2e suite already checks overflow and overlap. The work is subtraction plus one deliberate
visual language, not a rewrite of the product.

---

## 2. Decisions taken in the interview

| Question | Decision |
|---|---|
| Scope | Welcome, shared tokens/components, i18n copy — and, because tokens are inherited anyway, **all screens now**: guest, TV, admin included |
| Brand vs slop | Redesign the visual language rather than trim the neon |
| Deliverable | **`web/DESIGN.md` + automated gate first**; screens then migrate against it |
| Structural freedom | Any layout change that serves the task; routes, flows, and server contracts unchanged |
| Token blast radius | Global: the new language applies app-wide, each screen landing in its own change |
| Direction | **Concert poster** |
| Type | System stacks only, no webfonts — poster character from weight, case, tracking, size |
| Accent | New single accent, **acid yellow**; magenta/cyan/violet/lime retire |
| Poster limits | Poster everywhere, scaled per screen: display sizes on Welcome/TV, a calm but same-language utility scale on phone |
| Gate location | **Go architecture test** (`internal/architecture/`), runs on `go test ./...` |
| Banned by the gate | Gradients, glass/glow, arbitrary radii, off-scale type & spacing (warn), glyph/emoji icons, filler copy |
| Raw hex | Not banned: tokens are canonical in `styles/`, a view may use a literal one-off |
| Surfaces | **Dark stage + paper accents** (cream panels with dark ink) |
| Scales | **Snap to what already exists** — quantise current values, don't invent a new rhythm |
| DESIGN.md contents | Tokens, component inventory with states, per-screen audit, copy rules, review checklist, migration order |
| Gate strictness | Fail on bans; **warn** on scale drift (type/spacing) |
| Escape hatch | Not decided — see §10 open questions (proposal: one greppable inline comment) |
| E2E tests | Selectors and assertions may be updated; add `data-testid` |
| Motion | Functional only |
| Contrast | Automated check on token pairs, **advisory** while the palette settles |
| Doc position | `web/DESIGN.md`, referenced from the root `README.md`; the skill is *not* vendored |
| Copy voice | Minimal rewrite: remove filler and vague words, keep the existing specific wording |
| Copy parity | EN and FR rewritten together (see §6) |

---

## 3. The visual language: concert poster

### 3.1 Concept

A karaoke night looks like a printed gig poster taped to the bar door: huge condensed-looking
type, hard rules, one hot ink, paper stock, no glow. The app becomes that poster — dark where a
stage is dark, paper where a poster is paper.

Three rules that decide most cases:

1. **Type and rules carry the design.** Hierarchy comes from size jumps, uppercase, tracking and
   1px/2px rules. Never from light: no glow, no blur, no glass.
2. **One accent, used like ink.** Acid yellow marks the primary action, the current item and live
   state. It never fills a large surface, never appears twice in one block, and never sits on
   paper as text (contrast 1.02:1 — see §4.3).
3. **Everything is either stage or paper.** A screen region is dark (interactive, in a dark room)
   or paper (a printed notice: privacy, how-it-works, modals that ask a question). No mixing
   inside one card.

### 3.2 Surfaces

- **Stage (default):** near-black `--bg`, flat `--surface` blocks separated by 1px `--border`
  rules. Applies to TV, guest phone screens, admin, and the top of Welcome.
- **Paper:** `--paper` cream with `--ink` text, no radius beyond `--radius-sm`, 1px `--ink`
  border or a 2px top rule. Applies to at most one or two blocks per screen: Welcome's
  how-it-works and privacy notices, the nickname sheet, confirmation modals, the admin
  destructive-action dialog.
- Paper blocks never sit on top of each other, never contain the YouTube player, and never use the
  accent as text (only as a 2px rule or a 4px marker box).

### 3.3 Typography

System stacks only. The app is embedded in a Go binary and runs offline in a Docker image; a
webfont is a 30–60KB asset plus a licence plus FOUT, and the poster character is achievable
without it.

```css
--font: ui-sans-serif, system-ui, "Helvetica Neue", Helvetica, Arial, sans-serif;
--mono: ui-monospace, "SFMono-Regular", Menlo, Consolas, monospace;
```

Poster technique, applied through component classes rather than inline CSS:

| Role | Rule |
|---|---|
| Display (`--text-xl` and up) | `font-weight: 800`, `text-transform: uppercase`, `letter-spacing: -0.03em`, `line-height: 0.95`, `text-wrap: balance` |
| Section heading | `font-weight: 800`, uppercase at `--text-sm`–`--text-md`, `letter-spacing: 0.06em`, preceded by a 2px accent rule 32px wide |
| Body | sentence case, `--text-sm`, `line-height: 1.55`, max `68ch` |
| Meta / labels | `--text-xs`, uppercase, `letter-spacing: 0.08em`, `--muted` |
| Codes, durations, queue positions, PINs | `--mono`, uppercase, `letter-spacing: 0.12em` (promoted from decoration to a real role) |

The previous design used mono as an accent; now mono means "this is data you read or type".

### 3.4 Palette (measured, WCAG 2.1 ratios computed from these exact values)

| Token | Value | Use | Contrast |
|---|---|---|---|
| `--bg` | `#0B0B0C` | page base | — |
| `--surface` | `#141414` | raised block | — |
| `--surface-2` | `#1B1B1C` | input, secondary button | — |
| `--border` | `#2A2A2C` | 1px rules | — |
| `--border-strong` | `#3D3D3F` | focused/secondary borders | — |
| `--text` | `#F5F3EE` | primary text | 17.7:1 on `--bg`, 16.6:1 on `--surface` |
| `--muted` | `#B4B1A9` | secondary text | 9.2:1 on `--bg`, 8.0:1 on `--surface-2` |
| `--dim` | `#8A877F` | disabled, timestamps, placeholders | 5.5:1 on `--bg`, 4.8:1 on `--surface-2` |
| `--accent` | `#D8FF3E` | primary action, live state, current item | 17.1:1 on `--bg` |
| `--accent-ink` | `#0B0B0C` | text on accent | 17.1:1 |
| `--paper` | `#F2EDE3` | paper blocks | — |
| `--ink` | `#121110` | text on paper | 16.2:1 on paper |
| `--ink-muted` | `#5C5850` | secondary text on paper | 6.1:1 on paper |
| `--danger` | `#FF5A3C` | destructive, error | 6.4:1 on `--bg` |

Rules: accent on paper is **banned** (1.02:1). Error text uses `--danger` on stage and `--ink` with
a danger rule on paper. Selection color becomes accent-on-ink, not magenta-on-white.

### 3.5 Scales

**Type** — eight steps plus one fluid display (replaces 28 ad-hoc values):

| Token | Value | Absorbs current |
|---|---|---|
| `--text-2xs` | 0.75rem (12px) | 0.62, 0.65, 0.66, 0.7, 0.72, 0.75 |
| `--text-xs` | 0.875rem (14px) | 0.76, 0.78, 0.8, 0.82, 0.85 |
| `--text-sm` | 1rem (16px) | 0.86, 0.88, 0.9, 0.92, 1, 1.05, 1.08 |
| `--text-md` | 1.25rem (20px) | 1.15, 1.2, 1.25, 1.3, 1.35 |
| `--text-lg` | 1.5rem (24px) | 1.4, 1.5 |
| `--text-xl` | 1.875rem (30px) | 1.8 |
| `--text-2xl` | 2.5rem (40px) | 2.2, 2.4 |
| `--text-3xl` | 3.5rem (56px) | 4 |
| `--text-display` | `clamp(2.5rem, 7vw, 5rem)` | new; Welcome hero and TV now-playing only |

Body text grows from ~14.4px to 16px; that is deliberate (phone legibility in a dark room) and is
the one place where "snap to existing" rounds up rather than down. Line heights pair with the
ramp: `1.5` for body, `1.25` for `--text-md`/`--text-lg`, `0.95` for display.

**Space** — 4px grid, nine steps (replaces eleven+ ad-hoc gaps):

| Token | px | Absorbs |
|---|---|---|
| `--space-1` | 4 | 2, 3, 4px, 0.35rem |
| `--space-2` | 8 | 6, 7, 8px, 0.4–0.55rem |
| `--space-3` | 12 | 0.6–0.75rem, 10–12px |
| `--space-4` | 16 | 0.8–0.9rem, 1rem, 14px |
| `--space-5` | 24 | 1.3–1.5rem |
| `--space-6` | 32 | 2rem |
| `--space-7` | 48 | 3rem |
| `--space-8` | 64 | 4rem |
| `--space-9` | 96 | section gaps on Welcome/TV |

**Radius** — near-square, five tokens (replaces 14 values):

| Token | Value | Use |
|---|---|---|
| `--radius-0` | 0 | cards, paper blocks, queue rows, player |
| `--radius-sm` | 4px | buttons, inputs, badges |
| `--radius-md` | 8px | modals, sheets |
| `--radius-lg` | 12px | only the TV join panel and QR frame |
| `--radius-pill` | 999px | chips and toggles on an explicit allowlist only |
| `50%` | — | avatars, connection dots, spinners, ticket notches (§3.7) only |

**Elevation** — no blur shadows. Overlays get a 1px border plus a higher surface; the only shadow
token is a hard print offset:

```css
--shadow-hard: 4px 4px 0 var(--bg);   /* poster stamp, primary CTA hover/pressed only */
```

**Motion** — functional only: `spin` (spinners), `fade` 120ms (overlay/modal entrance), 100ms
color/transform transitions on buttons and rows. `pulse` and `rise` are removed. The global
`prefers-reduced-motion` override stays and must be the only place that zeroes durations.

### 3.6 Icons

No emoji and no typographic glyphs in UI chrome. One component, `web/src/ui/common/Icon.svelte`:
inline SVG, 24px grid, `stroke: currentColor`, 1.5px stroke, round caps, named set only.

Initial set (each added when first used, not speculatively): `search`, `plus`, `close`, `chevron`,
`shuffle`, `play`, `skip`, `up`, `down`, `lock`, `qr`, `volume`, `check`, `alert`.

Rules: a label beats an icon; icon-only buttons require `aria-label` and a 44px hit area; emoji is
allowed only inside user-generated content (nicknames, YouTube titles, channel names) and never in
product copy, toasts, or empty states.

### 3.7 Welcome layout — the reference implementation

The Welcome page is the one screen designed in full in this spec; every other screen follows its
tokens, components and rhythm. It deletes the hero glow, the gradient-clipped headline and the
three identical icon-title-sentence cards, and replaces them with a poster: a masthead, a split
hero (stage copy + paper ticket), a flat stage panel, a numbered line-up on rules, a paper
programme, a paper privacy notice, and an imprint footer.

**Kept, must not change:** the create-room action and its label (`common.createRoom` =
"Start a karaoke room" / "Créer une salle de karaoké" — `flow.mjs` and `audit.mjs` click it by
accessible name), the join form (input + submit + `role="alert"` errors), the language toggle, the
`#how` and `#privacy` anchors, the single `h1`, and the existing `welcome.*` i18n keys.

**Deleted:** `.glow-a` / `.glow-b` blurred orbs, the `background-clip: text` gradient headline,
the 3D-tilted stage mock with its fake waveform, the `.desktop-only` nav links, the `66vh` hero
min-height, and the three-card grid.

Wireframe, desktop ≥1024 (colours: plain = stage, `▓` = paper, `█`/`■` = accent):

```
┌───────────────────────────────────────────────────────────────────────────┐
│ FANA KARAOKE           HOW IT WORKS   PRIVACY               [EN | FR]     │
├─────────────────────────────────────────────────┬─────────────────────────┤
│ KARAOKE IN THE BROWSER                          │ ▓█ 2px accent           │
│                                                 │ JOIN A ROOM             │
│ TURN ANY SCREEN                                 │ ┌─────────────────────┐ │
│ INTO A KARAOKE                                  │ │ BLUE-MIC            │ │
│ ROOM.                                           │ └─────────────────────┘ │
│                                                 │ [       JOIN        ]   │
│ Put the room on the TV. Guests scan the code,   │ Ask the host for the    │
│ queue songs, and vote. You keep the controls.   │ code on the TV.         │
│                                                 │ - - - - - - - - - - - - │
│ [ START A KARAOKE ROOM ]   HOW IT WORKS ->      │ ■ NO ACCOUNTS           │
│                                                 │ ■ IN MEMORY ONLY        │
├─────────────────────────────────────────────────┴─────────────────────────┤
│ WHAT THE TV SHOWS                                          BLUE-MIC       │
│ ┌───────────────────────────────────────────────────────────────────────┐ │
│ │ BLUE-MIC                                             ■ LIVE            │ │
│ │ NOW SINGING                                                           │ │
│ │ BOHEMIAN RHAPSODY                                    ┌────────┐       │ │
│ │ Karaoke version                                      │  QR    │       │ │
│ │ NEXT · Dancing Queen                                 └────────┘       │ │
│ └───────────────────────────────────────────────────────────────────────┘ │
├───────────────────────────────────────────────────────────────────────────┤
│ THE LINE-UP                                                               │
│ 01  SEARCH     Find a karaoke track on YouTube and add it to the room.    │
├──────────────────────────────────────────────┬────────────────────────────┤
│ 02  VOTE                                     │ 03  JOIN                   │
│     Upvote what plays next. The host can     │     No accounts, no        │
│     still reorder the queue.                 │     downloads. Scan the    │
│                                              │     code on the TV.        │
├───────────────────────────────────────────────────────────────────────────┤
│ ▓▓ paper: HOW IT WORKS                                                    │
│ 01 Open this page on the   │ 02 Guests scan the  │ 03 Press play. The     │
│    screen plugged into the │    QR code and add  │    TV sings; you run   │
│    speakers.               │    songs.           │    the show.           │
├───────────────────────────────────────────────────────────────────────────┤
│ ▓▓ paper: PRIVACY                                                         │
│ PRIVACY  │ Rooms are stored in memory only. No accounts, no analytics…    │
│          │ Playback uses the official YouTube player.                      │
├───────────────────────────────────────────────────────────────────────────┤
│ FANA KARAOKE · THE ROOM IS THE WEBPAGE        CREATED BY FANAMPERANTSOA…  │
└───────────────────────────────────────────────────────────────────────────┘
```

Mobile 390: the same order in one column — masthead (brand + language toggle only), eyebrow, `h1`,
lede, full-width create button, paper ticket, stage panel, stacked line-up, paper programme
(stacked with 1px ink rules), paper privacy, footer.

#### Block specification

All values are tokens from §3.4–3.5; no literal colours, radii or sizes in the markup.

| Block | Structure | Type & measure | Rules |
|---|---|---|---|
| Shell | max-width 1180px; sections separated by 1px `--border`; `padding-block` `--space-6` mobile / `--space-7` desktop; `padding-inline` `--space-5` / `--space-6` | — | Radius 0 throughout; flat `--bg`; no page glow, no page gradient |
| Masthead | flex, 56px / 64px tall, 1px bottom rule | Brand 16px/800 uppercase, tracking .04em; nav + toggle 12px mono uppercase, tracking .1em | Brand mark: existing note SVG at 28px, stroke `--text`, both circles `--accent`, in a 40px square, 1px `--border`, `--radius-sm`. Nav links 44px hit area, 2px accent underline on hover/focus |
| Hero | 1px `--border` box, no radius; `grid-template-columns: minmax(0,1fr) 420px` ≥1024, one column below; the shared edge is the 1px rule | Left cell `padding: 40px` (≥1024) / `--space-5` (mobile) | No `min-height`; height comes from content |
| Eyebrow | 32×2px accent rule + text on one baseline | 12px mono uppercase, tracking .18em, `--muted` | No pill, no dot |
| H1 | one element, may wrap to 3 lines | `--text-display` with a Welcome override of `clamp(2.75rem, 6.2vw, 5.5rem)` (§10 Q3), weight 800, uppercase, `line-height: .88`, `letter-spacing: -.035em`, `max-width: 11ch`, `text-wrap: balance`, solid `--text` | Solid colour only: no gradient, no `background-clip`, no `-webkit-text-stroke` |
| Lede | one paragraph | `--text-sm`, 1.55, `max-width: 46ch`, `--muted` | Trim to ≤ 2 sentences at 390px |
| Actions | primary + secondary, side by side ≥420px, stacked below | 56px tall, uppercase 14px tracking .08em | Primary: accent fill + `--accent-ink`, `--radius-sm`, hover `translateY(-1px)` + `--shadow-hard` (no blur). Secondary: transparent, 1px `--border-strong`, chevron icon at 12px (SVG, not `→`) |
| Join ticket | paper block, full height of the hero cell; `padding: 32px` | Heading 20px/800 uppercase ink; field label 12px mono uppercase .12em `--ink-muted`; input 56px, mono 20px tracking .12em, `text-transform: uppercase`; submit 52px, uppercase 13px | `--paper` bg, `--ink` text, radius 0, `border-top: 2px solid var(--accent)`. Input 1px `--ink` border. Submit: `--ink` fill, `--paper` text. Accent is never used as text on paper |
| Ticket perforation + stub | dashed 1px `--ink-muted` rule `--space-4` under the button, with two 12px `--bg` notches centred on the left and right edges | Stub rows: 6px accent square marker + 11px mono uppercase `--ink-muted` | The notches are the only `50%` radius on the page. Two stub facts only (`welcome.ticketNoAccounts`, `welcome.ticketInMemory`) |
| Stage panel | 1px `--border`, radius 0; 40px caption bar with 1px bottom rule; screen `aspect-ratio: 16/9`, `max-height: 460px`, `--bg` inside 1px `--border`, padding `--space-5` | Caption 11px mono uppercase .16em `--muted` left, room code right; room code inside the screen 56px mono `--accent`; NOW SINGING label 11px mono `--muted`; title `clamp(1.5rem, 3vw, 2.5rem)` 800 uppercase `.95`, clamped to 2 lines; channel 14px `--muted`; NEXT row 11px label + 16px title | LIVE = 6px accent square + 12px accent text (no glowing dot). QR frame 88px, white, 1px `--ink`, radius 0, 10px mono caption. No waveform, no `transform`, no glow; keeps `aria-hidden="true"` |
| Line-up | Section label + heading; row 01 spans the width on `56px 18ch minmax(0,1fr)`; rows 02/03 sit side by side, split by a 1px vertical rule, each `40px 1fr` | Label 12px mono accent uppercase .16em; heading 28px/800 uppercase; numbers 12px mono `--accent`; row-01 title 24px/800 uppercase; 02/03 titles 20px/800; bodies `--text-sm` `--muted`, max 46ch | `padding-block: --space-5`, `border-top: 1px solid --border` between rows. No boxes, no tinted icon squares, no pills. Mobile: one column, number inline before the title, body below |
| Programme | paper block; heading + three steps, `1fr 1fr 1fr` split by 1px `--ink-muted` vertical rules | Heading 20px/800 uppercase ink; step number 11px mono `--ink-muted`; body 16px ink, 1.5 | 2px `--accent` rule on the block's top edge (a rule, not text). Keeps `id="how"` |
| Privacy | paper block; `300px 1fr` grid (stacked below 900) | Heading 20px/800 uppercase ink; body 16px ink, `max-width: 68ch`; YouTube note 12px `--ink-muted` above a 1px ink rule | Keeps `id="privacy"`. No card shadow, no radius |
| Footer | 1px `--border` top; three zones on desktop, stacked at 420px | 11px mono uppercase, `--muted` | Credit link underlined 1px in `--text`; cyan link colour retires |

#### What happens to the three cards

The facts survive; the container does not. No icons, no tinted 44px squares, no three equal
boxes: the mono numbers do the job the icons were doing. Copy stays a minimal edit (§6), and the
keys are reused so nothing else has to move.

| Key | Current | Proposed EN | Proposed FR | Why |
|---|---|---|---|---|
| `welcome.lineupTitle` (new) | — | The room, the queue, the votes | La salle, la file, les votes | A line-up needs a title that names the content, not a slogan |
| `welcome.feature1Title` | Search YouTube | Search | Chercher | One-word row label; the body says the rest |
| `welcome.feature1Body` | …and add it in seconds. | Find a karaoke track on YouTube and add it to the room. | Trouve une version karaoké sur YouTube et ajoute-la à la salle. | Drops the time claim |
| `welcome.feature2Title` | Vote the queue | Vote | Voter | Matches the other two rows |
| `welcome.feature2Body` | Upvote what you want to hear next. The host can take over the order anytime. | Upvote what plays next. The host can still reorder the queue. | Vote pour ce qui passe ensuite. L’hôte peut encore réordonner la file. | Drops “take over the order”, keeps the fact |
| `welcome.feature3Title` | Everyone joins | Join | Rejoindre | Row label |
| `welcome.feature3Body` | No accounts, no downloads. Just a QR code and a phone. | No accounts, no downloads. Scan the code on the TV and sing. | Pas de compte, pas de téléchargement. Scanne le code sur la TV et chante. | Replaces the slogan cadence with the actual action |
| `welcome.stageCaption` (new) | — | What the TV shows | Ce que montre la TV | Caption for the stage panel |
| `welcome.ticketNoAccounts` (new) | — | No accounts | Aucun compte | Ticket stub fact |
| `welcome.ticketInMemory` (new) | — | In memory only | En mémoire seulement | Ticket stub fact, accurate to the privacy text |

FR wording is a proposal and should be read by a native speaker before merge; FR must carry the
same facts as EN, not a shorter paraphrase.

#### Files

| File | Role | Budget |
|---|---|---|
| `ui/welcome/WelcomeView.svelte` | Shell, masthead, hero copy cell, section order, locale-invariant markup | ≈180 lines |
| `ui/welcome/WelcomeJoinTicket.svelte` | New: the paper join block (heading, field, submit, perforation, stub) — reusable in modals later | ≈110 lines |
| `ui/welcome/WelcomeLineup.svelte` | New: replaces `WelcomeFeatures.svelte`, which is deleted | ≈110 lines |
| `ui/welcome/WelcomeProgramme.svelte` | New: paper how-it-works | ≈90 lines |
| `ui/welcome/WelcomePrivacy.svelte` | New: paper privacy notice, split out of `WelcomeFeatures.svelte` | ≈80 lines |
| `ui/welcome/WelcomeStagePreview.svelte` | Rewritten flat: caption bar + screen, no tilt/glow/waveform | ≈130 lines |
| `ui/welcome/WelcomeFooter.svelte` | Restyled to the imprint, structure unchanged | ≈60 lines |
| `ui/common/Icon.svelte` | New: `chevron`, `alert` first; grows only as screens need icons | ≈70 lines |

All stay under the 300-line file gate and the 40-line function gate.

#### Acceptance checks for Welcome

- Devtools scan: zero `linear-gradient`, `radial-gradient`, `filter`, `backdrop-filter`,
  `text-shadow`; the only `box-shadow` is `var(--shadow-hard)` on the primary's hover.
- Radii present on the page: `0`, `--radius-sm` (mark, buttons, input, toggle) and the two ticket
  notches. Nothing is pill-shaped.
- No emoji or typographic glyph in markup; the arrow in the secondary action is an SVG icon.
- One `h1`, uppercase, wrapping to 3 lines at 1440 and at most 4 at 390, never clipped or
  overflowing its cell.
- 390×844: single column, ticket directly under the primary action, no horizontal overflow, no
  target under 44px, input 56px tall with the uppercase mono placeholder.
- 1440×900: the hero's two cells share exactly one 1px edge; the ticket's 2px accent rule sits
  flush with the top of the hero box; stage panel 460px or shorter.
- Keyboard: brand → How it works → Privacy → language toggle → create room → room input → join,
  with the 2px accent focus ring visible at every stop.
- Create room still lands on `/tv/<code>`; join errors still announce through `role="alert"`.
- Contrast, from the §3.4 table: lede 9.2:1, ticket hint 6.1:1 on paper, stub 6.1:1, footnote
  6.1:1 — all above AA.
- Accent never appears as text on paper; the two accent marks in the stub are 6px squares.
- `npm run check`, `npm run build`, `flow.mjs`, `audit.mjs` pass unchanged.

---

## 4. `web/DESIGN.md`

Placed next to the code it governs, referenced from the root `README.md` (one line, in the
architecture section). It is written **first**, before any screen changes, and becomes the
authority for visual decisions; `fana-karaoke-spec.md` keeps product/behavior rationale and gains a
pointer. Contents:

1. **Concept** — the poster paragraph and the three deciding rules (§3.1).
2. **Tokens** — palette table with measured contrast pairs, type ramp with the migration table,
   space, radius, elevation, motion. Every token documented with *what it is for*, not just a value.
3. **Component inventory** — for each: purpose, anatomy, sizes, every state (default, hover,
   active, focus-visible, disabled, loading, error), and a do/don't line.

   | Component | Notes for the new language |
   |---|---|
   | `.btn` primary | accent fill, `--accent-ink` label, uppercase tracking, radius-sm, no gradient |
   | `.btn` secondary | transparent, 1px `--border-strong`, accent border on hover |
   | `.btn` ghost / danger | ghost = text only; danger = 1px danger rule + danger text, never a filled red block |
   | `.card` | surface + 1px rule, radius 0; `--card--paper` variant |
   | `.field` | label uppercase xs, input 48px min, 1px rule, accent focus ring (2px, offset 2px) |
   | `.badge` | uppercase 2xs, radius-sm, 1px rule; accent variant only for "live"/"next" |
   | `.modal` | stage or paper, radius-md, no blur backdrop (solid 70% black scrim) |
   | `.toast` | bottom-center, surface, 1px accent left rule instead of glow |
   | queue row | poster list: 1px rules between rows, mono position number, accent marker on the current item |
   | room code / PIN | mono, uppercase, tracking 0.12em, the largest type on the TV top bar |

4. **Copy rules** — voice, banned phrases, EN/FR conventions (§6).
5. **Per-screen audit** — the table in §7, kept current as a checklist with a "done" column.
6. **Review checklist** — §9.
7. **Migration order** — §8.

---

## 5. Enforcement: the gate

Runs with `go test ./...` like the existing architecture gates (same package, same helpers:
`repoRoot`, `filesWithExt`, `relPath`). Two new files, each well under the repo's 300-line limit,
with functions under 40 lines:

### 5.1 `internal/architecture/ui_slop_test.go` — bans (failing)

| Rule | Match | Applies to |
|---|---|---|
| Gradients | `(linear\|radial\|conic)-gradient` | `web/src/**/*.svelte`, `web/src/styles/**/*.css` |
| Glass | `backdrop-filter`, `filter:\s*blur` | same |
| Glow | `box-shadow` not using a `--shadow-*` token; `text-shadow` at all | same |
| Radius | `border-radius` not `var(--radius-*)`, `0`, or `50%` | `web/src/ui/**/*.svelte` |
| Glyph icons | emoji ranges + `[✕✓⌕◎♪↑↓→←⌫✎⋯★☆⤨🔒🎤]` in markup text | `web/src/ui/**/*.svelte`, excluding `Icon.svelte` |
| Filler copy | banned phrase list, case-insensitive (below) | `web/src/application/i18n/*.ts` |

Banned phrases: `lorem ipsum`, `seamless`, `effortlessly`, `effortless`, `elevate`, `unlock`,
`supercharge`, `revolutionary`, `cutting-edge`, `game-changing`, `next-level`, `empower`,
`leverage`, `robust`, `intuitive`, `state-of-the-art`, `best-in-class`, `dive in`, `unleash`,
`harness`, `in today's`, `whether you're`, `take it to the next level`, `get started`,
`learn more`, `discover more`.

Deliberately **not** banned: hardcoded hex in components (decision in §2), `999px` inside
`tokens.css`, `filter: grayscale`/`brightness` on video, animated `opacity`.

### 5.2 `internal/architecture/ui_scale_test.go` — drift (warnings)

Reports, without failing:

- `font-size` literals in `web/src/ui/**/*.svelte` that are not `var(--text-*)` and not inside a
  `clamp()`.
- `gap` / `padding` / `margin` literals that are not `var(--space-*)`, `0`, `auto`, `50%`, `100%`,
  or `env(safe-area-inset-*)`.
- Radius literals outside the scale.

Implementation: collect violations per file, `t.Logf` a grouped report, and only fail when
`SLOP_STRICT=1` is set (so the team can burn the list down and then switch it on). The report is
also the migration worklist — the same sweep that produced §1's numbers.

### 5.3 `internal/architecture/ui_contrast_test.go` — advisory

Parses `web/src/styles/tokens.css`, computes WCAG 2.1 ratios for the documented pairs
(`--text/--bg`, `--text/--surface`, `--muted/--bg`, `--muted/--surface-2`, `--dim/--bg`,
`--accent/--bg`, `--accent-ink/--accent`, `--ink/--paper`, `--ink-muted/--paper`,
`--danger/--bg`) and logs each ratio, flagging anything under 4.5:1 (3:1 for ≥24px display text).
Advisory while the palette settles: fails only with `UI_CONTRAST_STRICT=1`.

### 5.4 Escape hatch

Proposal (needs confirmation, §10): one greppable inline comment,
`/* ui-slop-allow: <reason> */` on the violating line or the line above, reason required, no
file-level blanket. An allowlist of file+rule pairs lives in the spec table so the count stays
reviewable. The gate reports the number of active exemptions so it cannot quietly grow.

### 5.5 What the gate cannot see

Rendered layout, contrast in practice, focus order, touch targets, and the "does this look
generic?" judgment. Those stay in the review checklist (§9) and the Playwright suite.

---

## 6. Copy

**Voice: minimal rewrite.** Keep every string that is already specific and concrete; change only
filler, vague claims, and slogans. No exclamation marks. Labels name the action ("Add song",
"Create room", "Skip"), never "Submit", "Get started" or "Learn more". EN and FR are edited
together; FR is not a literal gloss of EN but must carry identical facts and the same length
budget.

Strings already fine and kept as-is: `welcome.joinHint`, `welcome.howStep1/2/3`,
`welcome.privacyBody`, `welcome.youtubeNote`, `welcome.createError`, `welcome.joinError`,
`admin.*` error strings, `room.code`/`room.scan`/`room.pin`.

Strings flagged in the audit:

| Key | Current | Problem | Direction |
|---|---|---|---|
| `welcome.eyebrow` | "Browser karaoke, zero installs" | slogan-shaped fragment | a plain fact: "Karaoke in the browser. Nothing to install." |
| `welcome.title` | "Turn any screen into a karaoke room." | keep, it is specific | unchanged |
| `welcome.subtitle` | "Put the room on the TV. Guests scan…" | good | trim to one sentence if it wraps past three lines at 390px |
| `welcome.feature1Body` | "…and add it in seconds." | time claim doing nothing | drop "in seconds" |
| `welcome.feature2Title/Body` | "Vote the queue" / "Upvote what you want…" | title is not a verb phrase like the others | "Vote on the queue" |
| `welcome.feature3Title/Body` | "Everyone joins" / "No accounts, no downloads. Just a QR code and a phone." | slogan cadence, "Just" | "No accounts, no downloads. Scan the code and sing." → verify against the actual join flow, which uses a code *and* a QR |
| `app.tagline` | "The room is the webpage." | cryptic | leave unless it appears in UI chrome, then replace with a factual line |
| `app.credit` | "Created by Fanamperantsoa Rakotoasimbola" | fine | unchanged |

Three-card grid: the *content* of the three features survives, but not as three identical
icon-title-sentence cards. The replacement is specified in §3.7 — a numbered line-up on 1px rules,
with row 01 full width and rows 02/03 split by a vertical rule. The copy stays load-bearing.

---

## 7. Per-screen audit (worklist)

Severity: **H** = reads as slop at a glance, **M** = inconsistency, **L** = polish.

| Surface | File | Findings | Sev |
|---|---|---|---|
| Tokens | `styles/tokens.css` | 4 accents, 2 glow shadows, 18px radius, purple-tinted surfaces | H |
| Components | `styles/components.css` | 4 gradients (primary button, card), 7 shadows w/ blur, pill `.btn`, pill `.badge`, glass backdrop, `--shadow-neon` on primary CTA | H |
| Base | `styles/base.css` | `.spinner` accent, magenta `::selection`, off-scale helper sizes | L |
| Welcome | `ui/welcome/WelcomeView.svelte` | gradient hero, glow orbs, gradient-clipped `h1`, 66vh hero, oversized radius — redesign in §3.7 | H |
| Welcome | `ui/welcome/WelcomeFeatures.svelte` | 3 identical icon-title-sentence cards (`⌕ ↑ ◎`), gradient icons' tinted squares — replaced by `WelcomeLineup` per §3.7 | H |
| Welcome | `ui/welcome/WelcomeStagePreview.svelte` | 4 gradients, 2 blurred shadows, fake waveform, 3D tilt — flat rewrite per §3.7 | H |
| Welcome | `ui/welcome/WelcomeFooter.svelte` | fine structurally; inherits tokens | L |
| Guest | `ui/guest/GuestHeader.svelte` | glass header, glow, glyph `✕` | M |
| Guest | `ui/guest/GuestNowPlayingCard.svelte` | 2 gradients, rounded-over content | M |
| Guest | `ui/guest/GuestVignettes.svelte` | gradient shimmer, shadow glow, glyph `♪` | M |
| Guest | `ui/guest/GuestSearchTab.svelte` | 999px input pill, glyph `⌕` in the field | M |
| Guest | `ui/guest/GuestQueueList.svelte` | 50% circles for vote buttons, off-scale type | L |
| Guest | `ui/guest/GuestHostBar.svelte` | glass bar | M |
| TV | `ui/tv/TvRoomView.svelte` | 5 gradients, blurred shadow, glyph `🎤`, glow on stage | H |
| TV | `ui/tv/TvTopBar.svelte` | 50% badge, off-scale mono | L |
| TV | `ui/tv/TvBottomBar.svelte` | 3 gradients, glass, shadow, glyph `⤨` | H |
| TV | `ui/tv/TvPlayer.svelte` | glass frame around the player | M |
| TV | `ui/tv/TvJoinPanel.svelte` | glass panel, glow QR frame | M |
| TV | `ui/tv/TvStageOverlays.svelte` | glass overlay, shadow | M |
| TV | `ui/tv/TvSoundPrompt.svelte` | glass + 2 shadows, `🔒` | M |
| TV | `ui/tv/TvSkipBanner.svelte` | gradient banner, glass, `⤨` | M |
| TV | `ui/tv/TvQueueDrawer.svelte` | blurred drawer shadow | L |
| Admin | `ui/admin/AdminView.svelte` | hardcoded hex, generic card stack, `✕`/`→` glyphs | M |
| Copy | `application/i18n/en.ts`, `fr.ts` | §6 table | M |

26 Svelte components in `web/src/ui` total.

---

## 8. Migration order

Every phase ships on its own and leaves the app working. The gate lands first so nothing new
becomes slop while the old screens are being converted (expect it to be red on many files at
first — that is why bans need an allowlist from day one, see §10).

1. **`web/DESIGN.md`** — the content of §3–§6 written for humans, plus the audit table.
2. **Gate** — the three tests, with the current violations allowlisted so `go test ./...` is green
   from the first commit. Allowlist count printed on every run.
3. **Tokens** — rewrite `tokens.css` (palette, type, space, radius, elevation, motion), add the
   component classes that carry poster typography (`.display`, `.eyebrow`, `.rule`, `.paper`).
   Old names (`--magenta`, `--cyan`, `--radius`) are removed in the same change; every use site is
   updated. This is the biggest single diff and the riskiest; `npm run check`, `npm run build` and
   the three e2e suites gate it.
4. **Shared components** — `components.css` rewritten against the inventory, `Icon.svelte` added,
   glyph call sites replaced.
5. **Welcome** — the reference implementation of the language; built to the §3.7 layout, copy
   edited per §6.
6. **Guest** — phone screens, utility scale, poster via rules and the mono data role.
7. **TV** — largest type, furthest viewing distance; display type here, but same accent discipline.
8. **Admin** — calmest surface; mostly deletion of decoration and the card grid.

Between phases, the app may mix new and old screens — accepted. No screen is "done" while it
still carries an allowlist entry.

---

## 9. Verification and review checklist

Automated:
- `go test ./...` — architecture gates (file size ≤300 lines, function ≤40 lines, imports) plus the
  three new UI gates.
- `gofmt -l .`, `go vet ./...`.
- `cd web && npm run check` — 0 errors, 0 warnings.
- `cd web && npm run build` — `web/dist` is committed, so rebuild it in the same change.
- `web/e2e/flow.mjs`, `audit.mjs`, `voting.mjs` — update selectors to `data-testid` where a class
  name is really a style hook; keep the assertions (no horizontal overflow, TV regions don't
  overlap, join panel on screen, reachable controls).
- Token contrast report from `ui_contrast_test.go`.

Manual, recorded in the final report (not claimed from reading code):
- 390×844 and 1440×900, both themes' surfaces (stage and paper).
- Keyboard-only pass of Welcome → create room → TV → guest join: visible focus at 2px/offset 2,
  logical order, no trap.
- Every state exists: empty library, loading search, search error with retry, empty queue, long
  YouTube title at 2 lines, 20-item queue, offline banner, `prefers-reduced-motion`.
- Touch targets ≥44px; contrast ≥4.5:1 for text, ≥3:1 for ≥24px.
- No emoji or glyph icon anywhere in chrome; no gradient, blur or glow in devtools.

---

## 10. Open questions

1. **Escape hatch.** Recommended: one inline `/* ui-slop-allow: reason */` comment, reason
   mandatory, count reported. Alternative: no exemptions at all, which makes the TV stage's
   theatrical needs harder. Needs a decision before the gate is written.
2. **Allowlist shape.** Per file+rule (reviewable, grows slowly) versus per line (precise, noisy).
3. **`--text-display` at 390px.** `clamp(2.5rem, 7vw, 5rem)` gives 40px on a phone; the poster
   direction may want a tighter `clamp` on Welcome's hero than on the TV. Two tokens or a
   responsive rule?
4. **Paper on mobile.** The §3.7 design puts three cream blocks on Welcome (join ticket, programme,
   privacy). Does that read as intentional at 390px in a dark room, or as a light-mode bug? Fallback:
   the programme block drops to stage styling, leaving the ticket and privacy on paper.
5. **Admin access to the palette.** With hex literals allowed in components, admin's data tables
   may drift back into one-off colors; should admin be tokens-only by convention in DESIGN.md?
6. **Accent on the player.** The YouTube iframe brings YouTube's own UI (red progress bar, white
   controls). How much space do we give it before the poster frame reads as a mismatch?
7. **`app.tagline`.** "The room is the webpage" — rewrite or remove it from UI chrome?
8. **Third-party content.** YouTube titles and channel names arrive in mixed case with emoji; the
   gate must not touch them. Confirm the display treatment (uppercase? no — that would mangle
   names; keep sentence case, clamp at 2 lines).

---

## 11. Definition of done

- `web/DESIGN.md` exists and every rule the gate enforces is stated there in words a person can
  follow, with the do/don't example for each.
- `go test ./...` includes the three gates; they are green on a clean tree with a printed,
  explicit exemption count (target: zero by the end of phase 8).
- No gradient, backdrop-filter, blur shadow, pill-by-default, emoji/glyph icon, or filler phrase
  in `web/src`.
- One accent: `--accent` is the only chromatic color in the palette besides `--danger`.
- All four surfaces (Welcome, guest, TV, admin) use the same tokens and component classes; no
  screen defines its own palette, radius or type sizes.
- E2E suites pass with `data-testid` hooks; `npm run check` and `npm run build` clean; `web/dist`
  rebuilt and committed.
- The final report names what was checked at 390px and 1440px, by keyboard, and in the rendered
  DOM — and what was not.

---

## 12. Implementation notes (what shipped)

All phases landed in one change. Tokens, shared components, and every screen were converted;
`WelcomeFeatures.svelte` was deleted and replaced by four focused welcome components.

### Gates as built

| File | Behaviour |
|---|---|
| `internal/architecture/ui_slop_test.go` | Fails on gradients, glass/blur, text-shadow, any `box-shadow` that is not a `--shadow-*` token, off-scale `border-radius` in `web/src/ui`, emoji/glyph characters in markup, and filler copy in the i18n files. Emits the exemption count every run. |
| `internal/architecture/ui_scale_test.go` | Reports `font-size` and spacing literals outside the token scales as warnings; fails when `SLOP_STRICT=1`. |
| `internal/architecture/ui_contrast_test.go` | Computes WCAG ratios for the ten documented token pairs; fails when `UI_CONTRAST_STRICT=1`. All ten clear 4.5:1 today (lowest: `dim` on `bg`, 5.48:1). |

Escape hatch implemented as planned: `/* ui-slop-allow: <reason> */` on the line or the line
above. Current exemption count: **0**.

### Deviations from this plan

1. **Banned phrase list narrowed.** Bare `unlock`, `empower`, `leverage` and `harness` produced
   false positives on real UI strings (“Unlock queue”, the PIN screen). The gate now targets
   marketing collocations (`unlock your`, `unlock the power`, `elevate your`, …). DESIGN.md carries
the implemented list.
2. **Welcome hero line-height 0.95, not 0.88.** At 88px in French, accented caps and descenders
   collided with the next line. DESIGN.md's display role already said .95; the screen now matches it.
3. **The join ticket carries both `.ticket` and `.paper`** so the paper surface is one shared class;
   the two perforation notches are the only `50%` radius on the page (allowlisted in DESIGN.md).
4. **Interactive sizes:** the language toggle buttons are 40px (44px including the container), and
   the footer credit link is a 44px target. Both were found by the rendered audit, not the code.
5. **E2E selectors:** `.lock-btn` became `[data-testid="unlock-host"]`; the French/English tab
   assertions in `audit.mjs` became case-insensitive because tabs are uppercase now. No assertion
   was weakened — same facts, new hooks.
6. **Icon set grew** by `link` and `phone` for the TV tool buttons, `microphone` for empty states.

### Verification actually run

- `gofmt -l .` clean, `go vet ./...` clean, `go test ./...` green including the three new gates:
  **0 ban violations, 0 scale drift, 0 exemptions**, all contrast pairs ≥ 4.5:1.
- `cd web && npm run check` — 0 errors, 0 warnings; `npm run build` — clean (dist rebuilt and
  committed).
- Playwright, live server on :8090 with a fresh SQLite library:
  `flow.mjs` 16/16, `audit.mjs` 9/9, `voting.mjs` 5/5.
- A temporary rendered audit (removed after the run) checked 28 assertions across Welcome at
  1440×900 and 390×844, the TV and the guest phone: no gradients, no `backdrop-filter`, no resting
  `box-shadow`, one uppercase `h1` that does not collide with the lede, two-cell hero on desktop /
  one column on phone, three paper blocks, three line-up rows, 56px fields, every visible target
  ≥40px, 2px solid focus ring, sensible tab order, no page errors. 28/28.

### Caveats

- `audit.mjs`'s “search prefers karaoke” check depends on live YouTube ordering and the cached
  library. It failed once on a warm library whose cached query returned a lyrics video first, and
  passed on a fresh library. Unrelated to this change (no search code was touched).
- No screenshot review was done: verification is DOM geometry, computed styles and the existing
  flows. A human pass at real viewing distance on a TV is still worth doing.
- FR strings are proposals and should be read by a native speaker before release;
  `room.hostUnlock*`, `admin.*` and other operational strings were left untouched.
