# Fana Karaoke — design system

This is the authority for visual decisions in `web/`. Product behavior lives in
`../fana-karaoke-spec.md`; the migration plan and the audit that produced this file live in
`../remove-ai-slop-spec.md`. When this file and a screen disagree, the screen is wrong.

The rule set is derived from the `no-ui-slop` practice: work from the product's own evidence, use
one set of choices, design every state, and check the rendered result before calling it done.

---

## 1. The language: concert poster

A karaoke night looks like a printed gig poster taped to the bar door: heavy type, hard rules, one
hot ink, paper stock. The app is that poster. Dark where a stage is dark, paper where a poster is
paper.

Three rules decide most cases:

1. **Type and rules carry the design.** Hierarchy comes from size jumps, uppercase, tracking and
   1px/2px rules. Never from light: no glow, no blur, no glass.
2. **One accent, used like ink.** Acid yellow marks the primary action, the current item and live
   state. It never fills a large surface, never appears twice in one block, and never sits on paper
   as text (1.02:1 — unreadable).
3. **Everything is stage or paper.** A region is dark (interactive, in a dark room) or paper (a
   printed notice). Never both inside one card.

### Do / don't

| Do | Don't |
|---|---|
| `border: 1px solid var(--border)` between blocks | `box-shadow` with blur to separate blocks |
| `background: var(--accent)` on the one primary action | gradient fills, glow around buttons |
| `text-transform: uppercase` for display and labels | `background-clip: text` gradient headlines |
| `var(--radius-sm)` (4px) or `0` on controls | 18–24px rounded cards, pills by default |
| Emoji only in user content (names, video titles) | emoji or `✕ ⌕ ♪ ◎` as UI icons |
| Solid `--paper` block with `--ink` text | `backdrop-filter` glass panels |
| Two or three real facts, in the product's words | identical icon-title-sentence card grids |

---

## 2. Tokens

All values live in `src/styles/tokens.css`. No screen defines its own palette, radius, type size
or spacing step. Hardcoded hex is allowed in a component only for a genuine one-off; the shared
CSS is tokens-only.

### Colour

| Token | Value | Use | Contrast |
|---|---|---|---|
| `--bg` | `#0B0B0C` | page base (stage) | — |
| `--surface` | `#141414` | raised block | — |
| `--surface-2` | `#1B1B1C` | input, secondary button | — |
| `--surface-3` | `#232325` | pressed / selected row | — |
| `--border` | `#2A2A2C` | 1px rules | — |
| `--border-strong` | `#3D3D3F` | secondary borders, focused fields | — |
| `--text` | `#F5F3EE` | primary text | 17.7:1 on bg, 16.6:1 on surface |
| `--muted` | `#B4B1A9` | secondary text | 9.2:1 on bg, 8.0:1 on surface-2 |
| `--dim` | `#8A877F` | disabled, timestamps, placeholder | 5.5:1 on bg, 4.8:1 on surface-2 |
| `--accent` | `#D8FF3E` | primary action, live, current item | 17.1:1 on bg |
| `--accent-ink` | `#0B0B0C` | text on accent | 17.1:1 |
| `--accent-soft` | `rgb(216 255 62 / 0.14)` | accent wash, never a fill for text blocks | — |
| `--paper` | `#F2EDE3` | paper blocks | — |
| `--ink` | `#121110` | text on paper | 16.2:1 |
| `--ink-muted` | `#5C5850` | secondary text on paper | 6.1:1 |
| `--danger` | `#FF5A3C` | destructive, error | 6.4:1 on bg |
| `--danger-soft` | `rgb(255 90 60 / 0.16)` | error blocks | — |

Accent on paper is banned as text. It may appear on paper only as a 2px rule or a 6px square
marker.

### Type

System stacks only: the app ships inside a Go binary and an offline Docker image.

```css
--font: ui-sans-serif, system-ui, "Helvetica Neue", Helvetica, Arial, sans-serif;
--mono: ui-monospace, "SFMono-Regular", Menlo, Consolas, monospace;
```

| Token | Size | Role |
|---|---|---|
| `--text-2xs` | 12px | labels, meta, ticket stub |
| `--text-xs` | 14px | secondary labels, queue meta |
| `--text-sm` | 16px | body |
| `--text-md` | 20px | card and section subtitles |
| `--text-lg` | 24px | section headings, line-up titles |
| `--text-xl` | 30px | screen headings |
| `--text-2xl` | 40px | now-playing title, numbers on the TV |
| `--text-3xl` | 56px | room code on the TV |
| `--text-display` | `clamp(2.5rem, 7vw, 5rem)` | Welcome hero only |

Roles:

- **Display** (`--text-lg` and up): `font-weight: 800`, uppercase, `letter-spacing: -0.03em`,
  `line-height: .95`, `text-wrap: balance`.
- **Section heading**: 800, uppercase, 2px accent rule of 32px above it.
- **Body**: sentence case, `--text-sm`, `line-height: 1.55`, max `68ch`.
- **Meta / labels**: `--text-2xs`, uppercase, `letter-spacing: 0.08em`, `--muted`.
- **Data** (room codes, PINs, durations, queue positions, timers): `--mono`, uppercase,
  `letter-spacing: 0.12em`. Mono means "read or type this", never decoration.

### Space

`--space-1` 4 · `--space-2` 8 · `--space-3` 12 · `--space-4` 16 · `--space-5` 24 · `--space-6` 32
· `--space-7` 48 · `--space-8` 64 · `--space-9` 96 (px).

### Radius

`--radius-0` 0 (cards, rows, paper) · `--radius-sm` 4px (buttons, inputs, badges) · `--radius-md`
8px (modals, sheets) · `--radius-lg` 12px (QR frame, join panel) · `--radius-pill` 999px (chips and
toggles, allowlist only).

`50%` is allowed only for: avatars, connection dots, spinners, ticket notches.

### Elevation and motion

No blur shadows. `--shadow-hard: 4px 4px 0 var(--bg)` is the only shadow and only on the primary
action's hover/pressed state.

Motion is functional: `spin` (spinners), `fade` 120ms (overlays), 100ms colour/transform
transitions. No pulse, no rise, no parallax. `prefers-reduced-motion` zeroes durations globally and
is the only place that does.

---

## 3. Components

All shared classes live in `src/styles/components.css`; every state below is required.

| Component | Anatomy | States |
|---|---|---|
| `.btn` | uppercase label, `--text-xs`, tracking .08em, min-height 44px (`--btn-lg` 54px), `--radius-sm`, 100ms transition | default, hover (`translateY(-1px)`), active, focus-visible (2px accent, 3px offset), disabled (50% opacity), loading (spinner + text) |
| `.btn-primary` | accent fill, `--accent-ink` text, `--shadow-hard` on hover | as above |
| `.btn-secondary` | transparent, 1px `--border-strong`, accent border on hover | as above |
| `.btn-ghost` | no fill, `--muted` text | as above |
| `.btn-danger` | transparent, 1px `--danger`, `--danger` text; never a filled red block | as above |
| `.btn-icon` | 44×44 square (not a circle), 1px border, icon only, `aria-label` required | as above |
| `.card` | `--surface`, 1px `--border`, `--radius-0` | — |
| `.card-paper` / `.paper` | `--paper`, `--ink`, 1px ink rule, `--radius-0` | — |
| `.field` + `.label` | label 12px uppercase `--muted`, gap 8px | — |
| `.input` | 48px min (56 on Welcome), `--surface-2`, 1px `--border`, `--radius-sm` | placeholder `--dim`, focus accent border + 2px ring, error `--danger` border |
| `.badge` | 12px uppercase, `--radius-sm`, 1px border | neutral; `.badge-hot` accent; `.badge-cyan` accent border; `.badge-live` accent text + square marker |
| `.connection-dot` | 8px dot + label, dot `50%` allowed | live (accent), connecting (accent, pulse removed → static + label), offline (`--danger`) |
| `.lang-toggle` | `--radius-pill` container, square inner buttons | active = `--surface-3` + `--text` |
| `.toast` | `--surface`, 1px border, 2px accent left rule, bottom-centre | enter = fade 120ms |
| `.modal-backdrop` | solid `rgb(11 11 12 / 0.82)` scrim, **no** backdrop-filter | — |
| `.modal` | `--surface`, 1px `--border-strong`, `--radius-md`, `--shadow-hard` | — |

**TV stage.** The top and bottom bars sit in normal flow with no borders; the video fills the space
between them and is never covered. The chrome is separated from the picture by its solid `--bg`
block, not by a rule.

**Buttons on paper.** `.paper .btn-secondary` and `.paper .btn-ghost` invert to `--ink`; the
accent-filled primary is the only button that keeps its fill. Without this, `.btn`'s light text
lands on cream at 1.02:1.

### Icons

`src/ui/common/Icon.svelte` only. Inline SVG, 24px grid, `stroke: currentColor`, 1.5px, round
caps. Names: `search`, `plus`, `close`, `chevron`, `shuffle`, `play`, `skip`, `up`, `down`, `lock`,
`qr`, `volume`, `check`, `alert`, `microphone`.

A label beats an icon. Icon-only buttons need `aria-label` and a 44px hit area. No emoji in product
copy, toasts or empty states.

---

## 4. Copy

Voice: short, literal, confident — a bar host, not a brochure. Sentence case for sentences,
uppercase only through CSS. No exclamation marks, no slogans, no adjectives doing no work.

Labels name the action: "Add song", "Start a karaoke room", "Join", "Skip". Never "Submit",
"Get started", "Learn more", "Discover".

Banned in `src/application/i18n/*.ts` (enforced by the gate): lorem ipsum, seamless, effortless,
elevate your, unlock your, unlock the power, supercharge, revolutionary, cutting-edge, game-changing,
next-level, empower your, leverage the power, robust, intuitive, state-of-the-art, best-in-class,
dive in, unleash, in today's, whether you're, take it to the next level, get started, learn more,
discover more.

Bare verbs like "unlock" are not banned: this app really does unlock a queue, and the PIN screen
says so.

EN and FR carry the same facts and the same length budget. FR is written, not glossed.

---

## 5. States

Every screen ships these, not just the happy path:

| State | Requirement |
|---|---|
| Empty | says what is missing and what to do next, in one sentence |
| Loading | spinner or skeleton, never a blank region; label what is loading |
| Error | what failed, in plain words, plus a retry control |
| Long text | 2-line clamp on titles, ellipsis, never overflow |
| Many items | the list scrolls inside its region; the page itself does not scroll sideways |
| Small screen | 390px works: no horizontal overflow, targets ≥44px, primary action first |
| Reduced motion | no entrance animation |

---

## 6. Review checklist

Run before calling any UI change done, and say what was actually checked:

- [ ] 390×844 and 1440×900 in the browser; no horizontal overflow at either.
- [ ] Keyboard-only pass: logical order, visible focus at every stop, no trap.
- [ ] Text contrast ≥4.5:1 (≥3:1 for text ≥24px), using the token pairs above.
- [ ] Touch targets ≥44px; inputs have labels.
- [ ] Every state in §5 exists for the screen that changed.
- [ ] Zero gradients, `backdrop-filter`, blur shadows, text-shadows; radii only from the scale.
- [ ] No emoji or glyph icons in chrome.
- [ ] `npm run check`, `npm run build`, `go test ./...`, and the Playwright flows pass.
- [ ] `web/dist` rebuilt, because it is committed.

The Go gates in `internal/architecture/ui_*_test.go` enforce the mechanical rules. They cannot see
layout, contrast in practice, focus order or whether a screen looks generic; that is this
checklist's job.
