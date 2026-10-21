# Fana Karaoke — Product & Technical Specification

- **Status:** Draft for review
- **Date:** 2026-10-21
- **Working directory:** `karaoke/` (repository root contains unrelated Go projects; this app is self-contained here)
- **Working title:** Fana Karaoke
- **One-line pitch:** A browser-based karaoke room. The host puts the room on a TV, guests scan a QR code, search YouTube karaoke videos from their phones, vote songs up, and the TV plays them in order.

---

## 1. Summary

Fana Karaoke is a web app inspired by [Syng.Rocks](https://syng.rocks/) and its [Flathub listing](https://flathub.org/fr/apps/rocks.syng.Syng). Syng splits the product into a queueing web frontend and a separate mpv playback client. Fana Karaoke deliberately collapses that into **one webpage with two views**:

- **TV view** — the room. Big YouTube player, room code, QR code, now playing, up-next, skip-vote status, and host controls.
- **Guest view** — a phone-first page reached by scanning the TV's QR code. Search, add songs, vote songs up, start/participate in a skip vote, and (with the PIN) use host controls.

There is no desktop client, no account, no database, and no saved library. Rooms live in memory and disappear when the host ends them, when they are abandoned for two hours, or when the server restarts.

The product is bilingual (English and French), designed for a public HTTPS deployment, and built as a single Go binary serving an embedded Svelte frontend over HTTP and WebSocket.

---

## 2. Context and constraints from the existing repository

The repository already contains Go services built in a **hexagonal (ports and adapters)** style (`kanban/`, `dnd/`). This project follows the same engineering conventions where they apply:

- Go 1.26, module-local package layout, manual constructor injection in `cmd/server`.
- Domain and use cases independent of HTTP, storage, and Svelte.
- SQLite is **not** used here: rooms are ephemeral and in-memory by explicit product decision.
- WebSocket via `github.com/coder/websocket`.
- No frontend build step in `kanban`; this project introduces one because the chosen stack is Svelte. The built assets are embedded into the Go binary with `go:embed`, so deployment remains one process.
- Tests drive the domain/use cases without a network, HTTP through `httptest`, WebSocket with a real client, and external services through fakes.
- Keep files and functions small, as in `kanban` (400 lines per file / 40 lines per function is a good default, though this spec does not make it a hard architectural gate).

The current `/Users/frakotoasimbola/project/karaoke` directory is empty. This document specifies a greenfield implementation.

---

## 3. Goals

1. A host can open the app, create a room in one click, and put the TV view on a big screen.
2. A guest can join in under 10 seconds by scanning a QR code, with no install, account, or app store.
3. A guest can search YouTube for a karaoke version of a song, see useful metadata, and add it to the shared queue.
4. The room can upvote queued songs; the queue is ordered by votes unless the host takes manual control.
5. The TV plays the current song through the official YouTube IFrame Player, auto-advances when it ends, and survives a refresh by resuming near the right position.
6. Guests can start a time-boxed skip vote; the host can skip immediately.
7. The host can pause, resume, skip, remove songs, reorder, lock submissions, clear the queue, rotate the PIN, and end the room.
8. Every viewer stays in sync through WebSocket snapshots, with a clear reconnect and stale-state story.
9. The interface works well on a TV at 3 m and on a phone at 30 cm, in English and French.
10. The app is safe enough to expose on the public internet with rate limits, host secrets, and strict input validation.

### Success criteria (v1)

- Two phones + one TV can complete: create → join → search → queue → vote → play → skip vote → host skip → end.
- A TV refresh mid-song resumes within **±5 seconds** of the server-known position.
- A guest page reconnect restores the queue and vote state without a manual reload.
- Search failures never block adding a song by YouTube URL (metadata may be reduced).
- All user-visible strings exist in EN and FR; no English leakage in the French UI outside user-generated content and YouTube titles.
- No accounts, no persistent personal data, no downloads.

---

## 4. Non-goals (explicitly out of scope for v1)

- Lyrics overlay, pitch scoring, vocal removal / stem separation, audio recording.
- Downloading YouTube media, extracting audio, or proxying video streams through the server. Playback is embedding only.
- User accounts, login, profiles, payments, subscriptions, analytics, or AI content moderation.
- Saved/reusable playlists, persistent play history, favorites, or a song library.
- Native mobile/desktop apps, offline mode, Chromecast/AirPlay integration.
- Multi-instance / horizontally scaled deployment. Room state is process-local.
- Public room directory or “browse rooms”. Rooms are reached by code or QR.
- Moderation queues, approval workflows, per-person queue caps, or duplicate blocking.
- Custom video upload, S3/local media sources, mpv playback.
- Karaoke-specific video filtering beyond the query suffix and the duration/live/Shorts filters.

---

## 5. Product decisions (interview log)

These are the binding decisions from the interview. Where two answers conflicted, the later, more specific answer wins; the resolution is stated in the rules.

| Area | Decision |
| --- | --- |
| Playback control | Host controls playback; guests can start a **skip vote**. |
| Song source | YouTube search from the phone, plus QR-code joining. |
| Stack | Go backend + Svelte frontend. |
| Persistence | Ephemeral, in-memory. |
| Screens | Two distinct views: TV and guest phone. |
| Host identification | Random 4-digit **PIN**, shown on the TV; host unlocks the phone with it. |
| Skip vote | Majority-based, time-boxed; final interpretation in §9.7. |
| Audio | Only the TV plays audio; phones do not play video/audio. |
| Auto-advance | 5-second countdown after a song ends; host can cancel. |
| Playback failure | Auto-skip with a room notice. |
| Queue caps | Unlimited per person; room-wide cap still applies. |
| Duplicates | Allowed. |
| Search backend | Public Invidious API (no YouTube Data API key), with cache and stale results. |
| Search quality | Automatically append `karaoke` to queries. |
| Room shutdown | Host ends the room; a 2-hour no-host cleanup is the safety net. |
| Guest identity | Nickname typed per song; no persistent display name. |
| Build shape | Single Go binary with embedded Svelte assets. |
| Playlist meaning | Live queue **plus upvotes** (queue is ordered by votes). |
| Skip-vote eligibility | Current queue owners / current singer only. |
| Visual direction | Neon karaoke stage: dark, magenta/cyan, big type. |
| Language | English + French toggle. |
| Queue upvotes vs FIFO | Resolved: queue is sorted by vote score; new songs start at the bottom; ties are FIFO. |
| Manual host reorder | Once the host reorders, **host order is final** for the rest of the room; vote sorting is suspended. |
| Host control surface | Both the TV page and the host's phone (after PIN unlock). |
| PIN | Random 4-digit, displayed on the TV. |
| Search filters | Exclude live streams, Shorts, and videos longer than 10 minutes. |
| Reconnect | Resume the current song at the server-known position. |
| Multiple TVs | Each playback screen plays independently; echo is accepted. |
| Upvotes | Any joined device; one vote per song per device; tap again to remove. |
| Out of scope | Lyrics/scoring/vocal removal, downloads, accounts/payments/AI moderation. |
| Landing page | Welcome page + create room + join code field. |
| Search backend detail | Invidious API. |
| Paste URL | Always available. |
| Host phone remote | Controls **all** playback screens in the room. |
| Room code | Readable word pair, e.g. `BLUE-MIC`. |
| Deployment | Public internet, with hardening. |
| TV access | Secret host link; the room code only opens the guest view. |
| Capacity | 25 guests / 100 queued songs per room. |
| Abandoned rooms | Deleted after 2 hours without a host connection. |

---

## 6. Personas and primary flows

### 6.1 Host at a party

1. Opens `https://fana.example/` on a laptop connected to the TV.
2. Presses **Create a room**.
3. Lands on `/tv/<secret>` with a large player, the join QR, room code `BLUE-MIC`, and host PIN `4821`.
4. Presses **Start karaoke** once (browser autoplay gesture). The room is idle until a song is queued.
5. Can leave the TV on that screen; the page is the host console.

### 6.2 Guest

1. Scans the QR with a phone camera.
2. Opens `/r/blue-mic`; no name prompt.
3. Searches “Bohemian Rhapsody”, sees filtered results with thumbnail, channel, duration.
4. Taps **Add**, types their nickname for that song (`Alex`), and sees it in the queue.
5. Taps the vote button on a friend's song, or **Vote to skip** the current one.
6. Watches the tally update live. The TV skips when the vote passes.

### 6.3 Host using a phone as remote

1. On `/r/blue-mic`, taps **Host controls**.
2. Enters the 4-digit PIN shown on the TV.
3. Gets a sticky host bar: pause/resume, next, queue moderation, lock, end room.
4. Pause affects every TV in the room.

### 6.4 Refresh / network recovery

1. TV reloads mid-song.
2. It fetches the room snapshot, loads the current YouTube video, seeks to `elapsed`, and plays.
3. If the browser blocks autoplay, a **Tap to resume** overlay appears.
4. Phone reconnects with exponential backoff and replaces its state from the next snapshot.

---

## 7. Information architecture and routes

| Route | Serves | Notes |
| --- | --- | --- |
| `/` | Svelte welcome page | Create room, join by code, language toggle. |
| `/r/{code}` | Svelte guest room | Sets a device cookie if missing. Case-insensitive code. |
| `/r/{code}/tv` | Not used | TV access is **not** available by code. |
| `/tv/{hostToken}` | Svelte TV view | Secret host link. Sets the host cookie. 404 on bad token. |
| `/api/*` | Go JSON API | Same origin only. |
| `/ws?room={code}` | WebSocket | Push snapshots and playback telemetry. |
| `/assets/*` | Vite-built static assets | Fingerprinted, long cache. |
| `/healthz` | Liveness | `200 {"status":"ok"}`. No room details. |
| `/qr.svg?code={code}` | QR image | Portable SVG generated server-side; the join URL is encoded. |

The SPA fallback serves `index.html` for `/`, `/r/*`, and `/tv/*`. `/tv/*` must pass through host-token validation before the fallback; assets and API paths never fall back to HTML.

Room codes are normalized to lowercase in URLs and shown uppercase in the UI: `/r/blue-mic` ↔ `BLUE-MIC`.

---

## 8. Domain model

### 8.1 Entities

```text
Room
  Code              string        // "blue-mic"
  HostToken         string        // 128-bit base64url secret, server-only
  PIN               string        // 4 digits, displayed on TV
  CreatedAt         time.Time
  HostLastSeenAt    time.Time
  QueueLocked       bool
  ManualOrder       bool          // host has reordered; vote sorting suspended
  Entries           []QueueEntry  // ordered as displayed
  NowPlaying        *NowPlaying
  SkipVote          *SkipVote
  Notices           []Notice      // newest first, max 20, in-memory only
  Sequence          uint64        // insertion order + broadcast sequence

QueueEntry
  ID                string        // "q_<random>"
  VideoID           string        // 11-char YouTube ID
  Title             string
  Channel           string
  ThumbnailURL      string
  DurationSeconds   int           // 0 if unknown
  Nickname          string        // snapshot, 1-24 chars
  DeviceID          string        // opaque device that submitted it
  Score             int           // derived from Votes
  Votes             map[string]bool  // deviceID -> true
  InsertedSeq       uint64

NowPlaying
  Entry             QueueEntry    // removed from Entries when it starts
  Status            "loading" | "playing" | "paused" | "countdown" | "error"
  StartedAt         time.Time     // zero when paused
  OffsetSeconds     float64       // elapsed at StartedAt (or at pause)
  DurationSeconds   int
  TvReportedSeconds float64       // last position reported by a TV
  CountdownEndsAt   time.Time     // set during a 5s advance
  Generation        uint64        // protects against stale TV events

SkipVote
  StartedAt         time.Time
  ExpiresAt         time.Time     // StartedAt + 60s
  EligibleDevices   []string      // frozen at start
  YesDevices        map[string]bool
  Threshold         int           // floor(len(EligibleDevices)/2)+1

Notice
  ID                string
  Code              string        // i18n key, e.g. "notice.songSkippedVote"
  Params            map[string]string
  CreatedAt         time.Time

Device
  ID                string        // 128-bit cookie value
  LastSeenAt        time.Time
  Connected         bool          // at least one WS open
  IsHostPhone       bool          // host cookie unlocked
```

### 8.2 Room state machine

```text
IDLE ──song added / host play──▶ PLAYING (after optional 5s countdown)
PLAYING ──pause──▶ PAUSED ──resume──▶ PLAYING
PLAYING ──ended/failed/skip──▶ COUNTDOWN(5s) ──▶ PLAYING (next) | IDLE (empty)
COUNTDOWN ──host cancel──▶ PAUSED/IDLE (hold on next song)
ANY ──host ends room──▶ ENDED (deleted, clients show closed screen)
```

### 8.3 Rules (authoritative)

1. **Queue order.** When `ManualOrder == false`, entries are sorted by `Score` descending, then `InsertedSeq` ascending. A new entry has score 0 and is therefore appended at the bottom among zero-vote entries. When `ManualOrder == true`, the stored order is authoritative until the room ends; votes still display and score but never reorder.
2. **Host reorder.** The host drags a song, or uses **Move up/down**. The client submits the complete ordered list of entry IDs. The server validates that it is a permutation of the current entries, saves it, and sets `ManualOrder = true` permanently. A notice explains that host order is locked.
3. **Votes.** Any device with a valid guest or host session may toggle one vote per entry. Votes are keyed by device cookie, so refreshing does not double-vote; clearing cookies does. Duplicate entries of the same video are separate vote targets.
4. **Vote sorting vs current song.** The current song is not in the queue and cannot be voted up or down. Only a skip vote can end it early.
5. **No per-person cap.** A guest can add any number of songs until the room-wide queue reaches 100. Submissions require a nickname per song.
6. **Duplicates allowed.** The same video ID can appear multiple times.
7. **Queue cap.** At 100 entries, additions fail with `queue_full`; the UI disables the add button and shows the cap.
8. **Guest cap.** At most 25 concurrent guest devices (WS connected within the last 2 minutes). Host TV and host phone do not count. A 26th guest sees `room_full` and can retry later.
9. **Nicknames.** Trimmed, 1–24 Unicode characters, stored and rendered as plain text. Empty names are rejected. No uniqueness requirement, no moderation, no profanity filter in v1.
10. **Song metadata.** Titles, channels, thumbnails, and durations are always taken from the server's metadata lookup, never trusted from the client, except that a pasted URL may fall back to YouTube oEmbed when Invidious is unavailable.
11. **Locked queue.** When locked, song additions are refused. Votes, skip votes, playback, and host actions continue. The UI shows “Queue locked by host”.
12. **Host authority.** The TV view is authenticated by the secret `hostToken`; the phone is authenticated by PIN. Both can perform every host action. The PIN is a courtesy boundary, not strong authentication (it is displayed in the room).
13. **Skip vote eligibility.** At vote start, eligible devices are:
    - devices owning at least one entry currently in the queue, plus
    - the device that submitted the current song,
    - restricted to devices seen within the last 2 minutes.
    The eligible set and threshold are frozen at start. Each eligible device gets one yes vote, which cannot be changed. A device that disconnects does not reduce the threshold. If the eligible set is empty, the vote cannot start.
14. **Skip vote threshold and timing.** Passes when `YesDevices >= floor(EligibleDevices/2) + 1`. The vote lasts 60 seconds. Passing ends the song immediately and starts the normal 5-second next-song countdown. Failing or expiring starts a 30-second cooldown before another skip vote can start. Only one vote is active at a time. The host can cancel the vote.
15. **Host skip.** The host's **Next** action ends the current song immediately, no vote and no cooldown; it posts a notice.
16. **Auto-advance.** After `ended`, a playback error, or a passed/host skip, the server enters `countdown` for 5 seconds, shows the next song, and starts it. The host can cancel the countdown, which holds on the next-song card until **Play next**.
17. **Empty queue.** After the last song, the room returns to `IDLE` with the join QR. Once the TV has been started at least once, the first song added while idle starts the same 5-second countdown; the host can cancel. Before the first start gesture, additions are queued and the TV keeps showing **Start karaoke**.
18. **First play.** Before the first playback, the TV shows a **Start karaoke** button. No YouTube playback is attempted before that user gesture. If songs are already queued, the first one starts after the gesture (with the usual countdown).
19. **Failure skip.** If the IFrame player reports an error (`2`, `5`, `100`, `101`, `150`), or reports the video unavailable, the server removes the current entry, posts a localized notice naming it, and schedules the next song. It does not retry.
20. **Resume.** The server is authoritative for elapsed time while playing: `elapsed = OffsetSeconds + (now - StartedAt)`. On pause, the elapsed value is frozen into `OffsetSeconds`. A TV loading a room computes the position from the snapshot and seeks once the player is ready. It reports progress every 5 seconds; if the TV is more than 2 seconds from the server estimate, the server accepts the TV's position (avoiding endless fights with buffering).
21. **Multiple TVs.** Each TV plays independently and follows server commands. A second TV opening mid-song seeks to the current position. Echo is accepted and documented. Every TV view requires the secret host link.
22. **Device identity.** The server issues an opaque, HMAC-signed `fana_device` cookie on first guest access. It is not a login and carries no personal data. Host phone sessions use a separate `fana_host_<code>` cookie.
23. **Room cleanup.** A room is deleted when:
    - the host ends it, or
    - no host heartbeat has been seen for 2 hours (safety net), or
    - the process restarts.
    A room whose TV was never opened is deleted 15 minutes after creation. There is no persistence.
24. **Room code collisions.** Generate two words from a curated list until an unused pair is found; fail creation after 20 attempts. Codes are case-insensitive.
25. **No cross-room leakage.** Snapshots, WS broadcasts, and notices are scoped to one room. The host token and PIN never appear in a guest snapshot.

---

## 9. Functional specification

### 9.1 Welcome page (`/`)

Desktop and mobile responsive, neon stage styling:

- Wordmark **FANA KARAOKE** and a one-line explanation.
- Primary button **Start a karaoke room** (`Créer une salle`).
- Secondary card **Join a room** with a single code input. Accepts `blue-mic`, `BLUE MIC`, or `BLUEMIC`; normalizes to `blue-mic`; invalid/unknown code shows an inline error, not a full error page.
- Language switch EN / FR in the header, persisted in `localStorage`.
- Small footer links: “How it works”, “Privacy” (static modal or section), and a note that playback uses YouTube.
- No room list, no login, no marketing carousel.

Creating a room calls `POST /api/rooms` and navigates to the returned `tvUrl`. Button is disabled while the request is in flight; failures show a retry message and `429` is translated as “Too many rooms created from this network, try again later.”

### 9.2 TV view (`/tv/{hostToken}`)

Purpose: be legible from the couch and function as the host console.

Layout (desktop/TV ≥ 1024 px):

- **Stage:** 16:9 YouTube iframe, as large as possible, centered on a near-black background with a subtle magenta/cyan glow.
- **Top-left:** room code in large monospace/display type (`BLUE-MIC`) and a “Scan to join” label.
- **Top-right:** QR code (SVG) encoding the absolute join URL; PIN shown below it as `Host PIN 4821`; language toggle; a “TV link” copy button.
- **Bottom bar:** now playing title, channel, singer nickname, elapsed/total time; next-up card with thumbnail; skip-vote status; host playback controls.
- **Overlay controls:** fade out after 3 seconds of no pointer movement; return on mouse move, touch, or keypress. Controls: play/pause, next, cancel countdown, mute/volume, fullscreen. Do not expose YouTube's native controls in the default view (`controls=0`), except where browser policy requires the player's own UI.
- **Right drawer (host):** queue with drag handles, remove buttons, clear, lock, rotate PIN, end room. Opens from a handle, never covers the video permanently.
- **Idle:** video area shows a neon placeholder with the QR/code and “Add a song to begin”, plus **Start karaoke** on first load.
- **Countdown:** full-width overlay “Next: {title} — {nickname}” with a large 5→1 counter and **Cancel**.
- **Skip vote:** banner under the video: “Skip vote — 4 of 7 needed”, a progress bar, remaining seconds, and host **Cancel vote**.
- **Ended room:** “This room has ended” with a **Create a new room** link.
- **Bad token:** generic 404 page; never reveal whether the room exists.

Mobile browser opening a TV link also works, but the layout prioritizes the guest view on small screens (the TV page shows a warning and a link to the guest view).

TV connectivity badge: a small dot (`Live` / `Reconnecting…`). It must not obscure the video.

### 9.3 Guest view (`/r/{code}`)

Phone-first, maximum width 640 px, sticky header:

- Header: room code, “Fana Karaoke”, connection dot, language toggle, and a host-lock icon (opens PIN sheet).
- Segmented control: **Search** | **Queue (n)**.
- **Search tab:**
  - Search field with 350 ms debounce, minimum 2 characters, clear button.
  - Results list: thumbnail (16:9), title (2-line clamp), channel, duration badge; disabled/hidden badge for live/Shorts/over-10-min items (they are filtered server-side and should not arrive).
  - Add button per result. Tapping it opens a compact “Who's singing?” sheet with nickname input, remembered **only for that submission** (not pre-filled from a previous song, though focusing the field is allowed). Submit calls the API; optimistic “Adding…” state.
  - Paste box always visible below search: “Or paste a YouTube link”, with a Resolve/Add flow and the same nickname sheet.
  - Empty query state: short instructions; no suggestions in v1.
  - Search unavailable: inline banner “Search is unavailable right now. Paste a YouTube link instead.” Stale cached results get “Showing cached results”.
- **Queue tab:**
  - Now playing card (thumbnail, title, singer, elapsed bar, skip-vote button/status).
  - Up-next list ordered exactly as the TV shows it: position, thumbnail, title, singer, score, vote toggle, and a “you” marker for the current device's submissions.
  - Vote button: heart/up-arrow with count, filled when voted, disabled when the entry is current or the device is not eligible for skip votes (not applicable to upvotes).
  - Skip-vote action: if no vote, **Vote to skip**; if active and eligible, **Vote yes** and live tally; if active and not eligible, read-only tally with explanation “Only people in the queue can vote”; during cooldown, disabled with a countdown.
  - Empty queue: friendly state pointing to Search.
  - Host-unlocked: sticky host bar with pause/resume, next, lock toggle, and a “Manage” sheet (remove, reorder, clear, end room, rotate PIN? PIN rotation stays TV-only to avoid locking out the TV).
- **Nickname sheet:** 1–24 characters, required, autoFocus, Enter submits, max length enforced, inline error on empty.
- **Notice toasts:** localized, from the last 20 room notices; `aria-live="polite"`; do not stack more than 3.
- No video/audio playback on the guest view. A thumbnail only.

### 9.4 Host unlock (phone)

- Tapping the host icon opens a 4-digit PIN keypad or numeric input.
- `POST /api/rooms/{code}/host/unlock` with `{pin}`.
- Success sets `fana_host_<code>` (HttpOnly, SameSite=Lax, Secure in production, 12 h) and unlocks host UI.
- Wrong PIN: generic “Wrong PIN”, 5 attempts per 10 minutes per IP+room, then a 15-minute lockout (`429` with `retry_after`). The TV is unaffected.
- The PIN is always visible on the TV in the join panel. The TV page itself is already host-authenticated by the secret link, so no PIN prompt is needed there.
- Host rotation: TV can generate a new PIN; existing phone host sessions remain valid. Old PIN stops working immediately.

### 9.5 Search

**Backend:** public Invidious API. No YouTube Data API key, by decision.

- Query normalization: trim, collapse whitespace, strip control chars, max 120 characters. Append ` karaoke` unless the query already contains `karaoke` (case-insensitive).
- Request: `GET {instance}/api/v1/search?q=<query>&type=video&region=<YOUTUBE_REGION>&page=1`.
- Try up to 3 instances from a shuffled, configurable list. Per-instance timeout 4 s. Mark an instance unhealthy for 60 s after a network error, 5xx, or invalid JSON.
- Normalize result: `videoId`, `title`, `author`, best thumbnail (prefer 320×180), `lengthSeconds`, `liveNow`, `type`.
- Filter: keep only `type == "video"`, `liveNow != true`, `lengthSeconds > 0 && <= 600`, and non-Shorts. Shorts heuristic: `lengthSeconds <= 60`; additionally discard results whose URL/source indicates `/shorts/`. Document this heuristic as approximate.
- Return at most 20 results in relevance order, with `stale` and `provider` metadata omitted from the client response except a boolean `stale`.
- Cache: normalized query key + region + language, 24 h TTL, max 1,000 entries, LRU eviction. On all-instance failure, serve a stale cache entry up to 7 days old with `stale: true`; otherwise return `503 search_unavailable`.
- Duplicate video IDs within a result set are collapsed, keeping the first.
- No autocomplete, filters UI, or infinite scroll in v1.
- Metadata lookup for adding a song: `GET {instance}/api/v1/videos/{id}` with a 7-day cache. If all instances fail, fall back to:
  - YouTube oEmbed: `https://www.youtube.com/oembed?url=https://www.youtube.com/watch?v={id}&format=json` for title/author/thumbnail, duration unknown (`0`), live unknown.
  - Reject only if both fail.
- Search is rate-limited per IP (see §12) but does not require a device cookie; the paste/resolve path does require a room.

**Paste URL parsing** accepts:

- `https://www.youtube.com/watch?v=ID`
- `https://youtube.com/watch?v=ID&...`
- `https://m.youtube.com/watch?v=ID`
- `https://youtu.be/ID?t=...`
- `https://www.youtube.com/shorts/ID` (parsed, then filtered as a Short once metadata is known)
- `https://www.youtube.com/embed/ID`

Only `youtube.com`, `www.youtube.com`, `m.youtube.com`, `music.youtube.com`, and `youtu.be` hosts are accepted. No redirects are followed for user-supplied URLs. Video IDs must match `^[A-Za-z0-9_-]{11}$`.

### 9.6 Queue and upvotes

- Adding: `POST /api/rooms/{code}/songs` with `{videoId?, url?, nickname}`. Exactly one of `videoId`/`url`. Server validates room, lock, cap, metadata, filters, then appends.
- A successful add posts a notice (`notice.songAdded`, params: nickname, title) and broadcasts a snapshot.
- The submitting device is attached to the entry for skip-vote eligibility; the nickname is only a label.
- Upvote toggle: `POST /api/rooms/{code}/songs/{entryId}/vote`; returns the new score and whether the device voted. One device–one vote; toggle removes. No notification for votes (too noisy).
- Scoring is denormalized for display; recompute from the map on mutation.
- Host removal: `DELETE /api/rooms/{code}/songs/{entryId}`; posts a notice if removed from the queue. Removing the current song is done via **Next**, not the queue.
- Host reorder: `PUT /api/rooms/{code}/order` with `{entryIds:[...]}`; must be an exact permutation. Sets `ManualOrder`.
- Host clear: `POST /api/rooms/{code}/queue/clear`; removes all pending songs, keeps now playing; posts notice; requires confirmation.
- Host lock: `POST /api/rooms/{code}/queue/lock` `{locked:true|false}`; posts notice.
- Queue slice in snapshots includes `position` (1-based), `score`, `votedByMe`, `mine`, and `canVote` (always true for upvotes).

### 9.7 Skip vote

Trigger: guest taps **Vote to skip** on the now-playing card.

Server flow:

1. If a vote is active, return its state (idempotent).
2. If in cooldown, return `409 vote_cooldown` with `retry_after`.
3. Compute eligible devices (queue owners + current singer, seen ≤2 min).
4. If none, return `409 no_eligible_voters`.
5. Freeze the list, compute `threshold = floor(n/2)+1`, set `ExpiresAt = now+60s`, broadcast.
6. The initiator's vote is **not** automatically cast; they must tap **Vote yes** (or the UI can auto-cast with an explicit confirm — spec chooses explicit tap to avoid accidental skips).
7. Each eligible device may cast one yes; duplicates are ignored. After each cast, broadcast the tally.
8. When `yes >= threshold`, end the current song immediately, record notice `notice.songSkippedVote` (yes, eligible, nickname of current singer, title), clear the vote, and start the 5 s countdown.
9. At expiry, clear the vote, start 30 s cooldown, notice `notice.skipVoteFailed` with the tally.
10. Host **Cancel vote** clears it without cooldown; host **Next** skips directly and clears any vote.
11. If the song ends naturally during a vote, discard the vote. If the current song changes for any reason, discard the vote.
12. Non-eligible guests see the tally but no vote button, plus an explanation.

Snapshot shape:

```json
"skipVote": {
  "active": true,
  "yes": 3,
  "threshold": 5,
  "eligible": 9,
  "expiresAt": "2026-10-21T20:15:00Z",
  "votedByMe": false,
  "canVote": true,
  "cooldownUntil": null
}
```

### 9.8 Playback

**TV client** uses the YouTube IFrame Player API, loaded from `https://www.youtube.com/iframe_api`.

Player parameters:

```text
autoplay=0
controls=0
disablekb=1
fs=1
iv_load_policy=3
modestbranding=1        // deprecated but harmless; rely on CSS overlay
playsinline=1
rel=0
origin=<PUBLIC_BASE_URL>
```

No custom proxy, no `youtube-nocookie` requirement in v1 (configurable later). The player container is always visible; hiding the video to play audio-only is not allowed.

Commands and telemetry:

- Server snapshot contains `nowPlaying` with `videoId`, status, offset estimate, duration, generation.
- TV loads the video when `entryId` changes. On `onReady`, seek to the estimated position if > 2 s, then play if status is playing.
- If `playVideo()` is rejected/blocked, show **Tap to resume**; do not mark the song failed and do not skip.
- `onStateChange(ENDED)` sends `{type:"tvEnded", entryId, generation}` over WS. The server accepts only the first end for the current generation, enters countdown, and broadcasts.
- Player errors send `{type:"tvError", entryId, code}`; server auto-skips per rule 19.
- Every 5 s while playing, send `{type:"tvProgress", positionSeconds, durationSeconds, generation}`.
- On `PAUSED`/`PLAYING` initiated from the TV's own controls, send an action to the server so phones reflect the change. TV-initiated pause uses `POST /playback {action:"pause"}`.
- Pause broadcasts a command; every TV pauses. Resume broadcasts; every TV seeks to the stored offset if it drifted and plays.
- Next: server removes current, posts notice, clears vote, starts countdown.
- The server owns the countdown deadline; TVs display the deadline from the snapshot. A late-joining TV shows the remaining time, not a fresh 5 s.
- During countdown, TVs show the next card and pause at 0. The server starts the next entry at the deadline. If the queue is empty, go idle.
- Idle: no video loaded; a branded panel instead.
- Song duration is the metadata duration; if it is 0, the TV reports the actual duration from the player and the server stores it.

Elapsed math details:

- `playing`: `elapsed = OffsetSeconds + now.Sub(StartedAt)`.
- `paused`: `elapsed = OffsetSeconds`.
- `loading`: `elapsed = OffsetSeconds` (usually 0).
- On resume: `OffsetSeconds = elapsed(at pause)`, `StartedAt = now`.
- On seek / TV drift correction: `OffsetSeconds = reported`, `StartedAt = now`.
- Clamp elapsed to `[0, DurationSeconds]` when duration > 0.

### 9.9 Host controls (summary)

| Action | TV | Phone after PIN | Endpoint / message |
| --- | --- | --- | --- |
| Play / resume | Yes | Yes | `POST /playback {action:"play"}` |
| Pause | Yes | Yes | `POST /playback {action:"pause"}` |
| Next | Yes | Yes | `POST /playback {action:"next"}` |
| Cancel countdown | Yes | Yes | `POST /playback {action:"cancelCountdown"}` |
| Remove queued song | Yes | Yes | `DELETE /songs/{id}` |
| Reorder queue | Yes (drag) | Yes (move buttons) | `PUT /order` |
| Clear queue | Yes | Yes | `POST /queue/clear` |
| Lock / unlock queue | Yes | Yes | `POST /queue/lock` |
| Cancel skip vote | Yes | Yes | `POST /skip-vote/cancel` |
| Rotate PIN | Yes | No | `POST /pin/rotate` |
| End room | Yes | Yes (confirm) | `POST /end` |
| Copy TV link | Yes | No | Client-side from known URL only on TV |
| Mute / volume / fullscreen | Yes | No (controls all TVs for playback; volume is local) | Local only |

All host endpoints require the host cookie and return `403 host_required` otherwise. Every host action broadcasts a snapshot; destructive actions post notices.

### 9.10 Notices and feedback

Notices are ephemeral, newest first, max 20, and included in every snapshot. They are **codes plus params**, localized on the client; the server never sends English prose to the UI.

| Code | When | Params |
| --- | --- | --- |
| `notice.songAdded` | Song added | `nickname`, `title` |
| `notice.songSkippedHost` | Host next/skip | `title`, `nickname` |
| `notice.songSkippedVote` | Vote passed | `title`, `yes`, `eligible` |
| `notice.skipVoteFailed` | Vote expired | `yes`, `threshold` |
| `notice.skipVoteStarted` | Vote started | `nickname` |
| `notice.songFailed` | Player error | `title`, `reason` (`unavailable`/`embed`/`error`) |
| `notice.songRemoved` | Host removes | `title` |
| `notice.queueCleared` | Clear | `count` |
| `notice.queueLocked` | Locked | — |
| `notice.queueUnlocked` | Unlocked | — |
| `notice.hostUnlocked` | Phone host unlock | — |
| `notice.orderLocked` | Host reorder | — |
| `notice.roomEnded` | End | — |

Toasts on phones; a compact notice ticker or toast on the TV (never over the video for more than 4 seconds). Critical state (failed song, vote result) is visible in the now-playing area as well.

### 9.11 Internationalization

- Languages: `en`, `fr`. Default from `navigator.language` (`fr*` → French), otherwise English; persisted in `localStorage["fana.lang"]`.
- Toggle in every view header; updates `<html lang>` and all strings immediately without reload.
- Message catalogs: `web/src/lib/i18n/en.ts` and `fr.ts`, typed, flat keys, interpolation with `{name}`.
- Server-facing errors use stable codes (`room_not_found`, `queue_full`, etc.); the client maps codes to localized messages. Unknown codes fall back to a generic message plus the HTTP status.
- User content (titles, nicknames, channel names) is never translated.
- Dates/times via `Intl.DateTimeFormat`; numbers via `Intl.NumberFormat`.
- Room-code words are English regardless of UI language.
- French copy must be natural, not machine-literal; use “salle”, “file d'attente”, “passer”, “vote pour passer”, “chanteur/chanteuse” neutrally where possible.

### 9.12 Room lifecycle

- Create → host token + PIN + code generated atomically; the room is inserted in memory and the host is redirected to the TV view.
- The TV page opens a WS as a host connection and heartbeats every 20 s; host phone sessions also heartbeat.
- `HostLastSeenAt` updates on any host WS heartbeat or host API request.
- Cleanup goroutine every minute:
  - delete rooms whose `HostLastSeenAt` is older than 2 h;
  - delete never-opened rooms 15 min after creation;
  - enforce `MAX_ROOMS` by refusing creation (`503 server_full`), not by evicting live rooms.
- Host **End room** immediately deletes the room, broadcasts `roomEnded` to all sockets, closes them after the message, and shows the ended screen on every client. The action requires a confirmation dialog on the phone and a hold-to-confirm / two-step confirm on the TV.
- Server shutdown: broadcast `serverRestarting`, wait up to 2 s, close sockets, exit. Clients show “The server restarted; this room is gone” and offer to create a new room.
- The server stores no history after a room is deleted; logs contain only room code, counts, and error codes, never nicknames or host tokens.

---

## 10. Screens and component breakdown (Svelte)

### 10.1 Frontend stack

- **Svelte 5** with runes (`$state`, `$derived`, `$effect`) and TypeScript.
- **Vite** with `@sveltejs/vite-plugin-svelte`; no SvelteKit, because Go owns routing, cookies, and SSR-free static delivery.
- `svelte-check` for type checking.
- **Vitest** + `@testing-library/svelte` for component/unit tests.
- **Playwright** for multi-context end-to-end tests.
- No CSS framework. Global design tokens + scoped component styles.
- No external fonts or icon CDN; inline SVG icons and a local/system font stack.

### 10.2 Directory shape

```text
karaoke/
  cmd/server/main.go
  internal/
    domain/              room.go, queue.go, vote.go, playback.go, errors.go
    ports/               input.go, output.go
    usecase/             room_service.go, queue_service.go, playback_service.go,
                         vote_service.go, search_service.go, cleanup.go
    adapters/
      primary/http/      handler.go, api_room.go, api_queue.go, api_playback.go,
                         api_search.go, auth.go, ws.go, static.go, qr.go,
                         middleware.go, errors.go
      secondary/memory/  room_repo.go, device_repo.go
      secondary/invidious/ client.go, cache.go, oembed.go
    events/              bus.go
  web/
    embed.go             //go:embed all:dist
    package.json
    vite.config.ts
    tsconfig.json
    index.html
    src/
      main.ts
      App.svelte
      lib/router.ts
      lib/api.ts
      lib/ws.ts
      lib/session.ts
      lib/i18n/{index.ts,en.ts,fr.ts}
      lib/stores/{room.svelte.ts,toast.svelte.ts}
      components/{Button,Modal,Toast,LanguageToggle,ConnectionDot,...}.svelte
      views/{Welcome,GuestRoom,TvRoom,NotFound,RoomEnded}.svelte
      features/search/{SearchPanel,ResultRow,NicknameSheet,PasteBox}.svelte
      features/queue/{QueueList,QueueEntryRow,NowPlayingCard,SkipVotePanel,VoteButton}.svelte
      features/host/{HostBar,HostUnlockSheet,QueueManager,EndRoomDialog}.svelte
      features/tv/{Player,PlayerOverlay,JoinPanel,NextUp,CountdownOverlay,TvControls}.svelte
    static/              favicon, og image
  Dockerfile
  Makefile               // build/dev/test targets
  README.md
  fana-karaoke-spec.md
```

### 10.3 State and realtime client

- One `RoomStore` per page, created from the initial `GET /api/rooms/{code}` snapshot.
- WebSocket is **push-only for state**; all mutations go through HTTP. This keeps authorization and validation in one place and makes retries explicit.
- On each snapshot, replace the room state wholesale keyed by `seq`; ignore older/duplicate seq. The payload at 100 entries is small enough (<50 KB) and broadcasts are coalesced server-side at 100 ms.
- On reconnect: exponential backoff from 500 ms to 10 s with jitter, unlimited retries while the page is visible; on `visibilitychange` to visible, reconnect immediately and refetch the snapshot.
- A socket close shows **Reconnecting…**; the UI remains readable but mutations are disabled until reconnected.
- localStorage holds only `fana.lang`; all identity is cookie-based.
- No optimistic queue mutations except the add button spinner. Playback controls wait for the server snapshot to avoid conflicting commands.

### 10.4 Design system

**Direction:** neon karaoke stage, dark-only, high contrast, party energy without hurting legibility.

Tokens (starting point, refinable):

```css
--bg: #08050F;
--surface: #120A1F;
--surface-2: #1B0F2E;
--border: #2E1B4A;
--text: #F7F2FF;
--muted: #B9A7D6;
--magenta: #FF2D95;
--cyan: #22D3EE;
--violet: #8B5CF6;
--lime: #B6FF3B;
--danger: #FF5C5C;
--radius: 18px;
--shadow-neon: 0 0 24px rgb(255 45 149 / 0.35);
```

- Display headings: `ui-sans-serif, "Helvetica Neue", Arial, sans-serif`, weight 800–900, tight tracking, uppercase for room codes.
- Body: same stack, 16–18 px on phone, 20–24 px on TV.
- Room code: tabular/monospace stack (`ui-monospace, "SFMono-Regular", Menlo, monospace`), huge, letter-spaced.
- Buttons: solid magenta primary, cyan outline secondary, 44 px minimum touch target, visible focus ring.
- Cards: dark violet surfaces, 1 px border, 18 px radius; no glassmorphism that harms contrast.
- Motion: subtle glow/pulse on vote and countdown; all nonessential motion disabled under `prefers-reduced-motion`.
- TV-specific scale: base font 1.25× desktop; controls and QR at least 96 px wide; no text below 18 px.
- Guest: bottom-safe-area padding, sticky segmented control, search input 48 px tall, no horizontal scroll at 320 px.
- Accessibility: WCAG 2.1 AA contrast; semantic buttons; `aria-live` for now playing, notices, vote tally, countdown; focus trap in dialogs; Esc closes; keyboard shortcuts on TV (`Space` play/pause, `N` next, `F` fullscreen, `?` help) documented in a small overlay.
- No custom cursor, no autoplay sound on the welcome page, no confetti by default.

---

## 11. Backend architecture

### 11.1 Hexagonal layering

- `domain`: pure entities and rules; imports only the standard library. No HTTP, JSON tags optional but no adapters.
- `ports`:
  - Input: `RoomService`, `QueueService`, `PlaybackService`, `VoteService`, `SearchService`, `CleanupService`.
  - Output: `RoomRepository`, `VideoCatalog`, `Clock`, `IDGenerator`, `Broadcaster`.
- `usecase`: orchestration, authorization checks at the application boundary (host token/cookie resolved by the HTTP adapter, passed as an actor), notices, state transitions, timers.
- `adapters/primary/http`: routing, cookies, JSON, WS, static assets, rate limiting.
- `adapters/secondary/memory`: map + per-room mutex; no global lock for reads.
- `adapters/secondary/invidious`: HTTP client, instance rotation, cache, oEmbed fallback.
- `cmd/server`: config from env, wiring, graceful shutdown.

Dependency rule: domain imports nothing internal; use cases import domain and ports; adapters import the core; the core never imports an adapter.

### 11.2 Concurrency

- One `sync.Mutex` per room for mutations; snapshots built under the lock and broadcast after release.
- A hub maps room code → set of sockets with a mutex; broadcast is non-blocking with per-socket buffered channels (size 16). Slow clients are dropped and reconnect.
- Timers: a single scheduler goroutine with a min-heap of deadlines (countdown end, vote expiry, cleanup), or per-room `time.AfterFunc` guarded by a generation token. Choose the scheduler for testability with a fake clock.
- Playback generation prevents stale TV events from advancing a new song.
- No goroutine may hold a room lock while writing to a socket.
- Race detector is part of the test suite.

### 11.3 Configuration

| Env var | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | HTTP listen port. |
| `PUBLIC_BASE_URL` | derived from request in dev | Absolute origin used for QR/join URLs, YouTube `origin`, cookies. Required in production. |
| `INVIDIOUS_INSTANCES` | small dev list | Comma-separated base URLs; rotated and health-tracked. |
| `YOUTUBE_REGION` | `FR` | Invidious region for search. |
| `SESSION_SECRET` | random at boot in dev | HMAC key for device/host cookies. Required and stable in production (restart invalidates sessions, acceptable). |
| `ROOM_TTL` | `2h` | No-host cleanup. |
| `MAX_ROOMS` | `1000` | Global room cap. |
| `MAX_GUESTS` | `25` | Concurrent guest devices per room. |
| `MAX_QUEUE` | `100` | Pending entries per room. |
| `SEARCH_CACHE_TTL` | `24h` | Query cache. |
| `SEARCH_STALE_TTL` | `168h` | Max stale cache age on failure. |
| `LOG_LEVEL` | `info` | `debug|info|warn|error`. |
| `TRUST_PROXY` | `false` | Trust `X-Forwarded-For` for rate limits only when behind a known proxy. |
| `DEV` | `false` | Relax Secure cookies, allow Vite dev proxy. |

Config is validated at startup; production refuses to start without `PUBLIC_BASE_URL` and `SESSION_SECRET`, and warns if `INVIDIOUS_INSTANCES` is empty.

---

## 12. HTTP API

All JSON. Errors use:

```json
{"error":{"code":"queue_full","message":"The queue is full.","retry_after":null}}
```

`message` is English for logs/debugging; clients localize by `code`. Unknown fields are ignored. Mutations require a same-origin `Origin` header (or no Origin for non-browser clients in dev) and a valid cookie. Content-Type must be `application/json` except `DELETE` without a body.

### 12.1 Public

| Method | Path | Body / query | Response |
| --- | --- | --- | --- |
| `GET` | `/api/config` | — | `{maxGuests, maxQueue, languages:["en","fr"], region}` |
| `POST` | `/api/rooms` | `{}` | `201 {code, joinUrl, tvUrl, pin}`. Rate-limited by IP. |
| `GET` | `/api/rooms/{code}` | — | Guest snapshot. Sets `fana_device` if absent. |
| `GET` | `/api/rooms/{code}/qr.svg` | `?base=` optional override? No; use `PUBLIC_BASE_URL` | `image/svg+xml` QR of `joinUrl`. |
| `GET` | `/api/search` | `q`, `lang` | `{results:[...], stale:false}`. Rate-limited. |
| `GET` | `/healthz` | — | `{"status":"ok"}`. |

`POST /api/rooms` response example:

```json
{
  "code": "blue-mic",
  "joinUrl": "https://fana.example/r/blue-mic",
  "tvUrl": "https://fana.example/tv/7f3k...Q",
  "pin": "4821"
}
```

The `tvUrl` is returned once to the creator and is also retrievable from the TV page for copying. It is **never** in a guest snapshot.

### 12.2 Guest (device cookie required)

| Method | Path | Body | Notes |
| --- | --- | --- | --- |
| `POST` | `/api/rooms/{code}/songs` | `{videoId}` or `{url}`, plus `{nickname}` | Validates lock/cap, resolves metadata, filters live/Shorts/>10 min. `201` with the entry. |
| `POST` | `/api/rooms/{code}/songs/{entryId}/vote` | `{}` | Toggle upvote. `200 {score, voted}`. |
| `POST` | `/api/rooms/{code}/skip-vote` | `{}` | Start a vote. Returns vote state. Idempotent if active. |
| `POST` | `/api/rooms/{code}/skip-vote/cast` | `{}` | Cast yes. `409 not_eligible` / `409 no_vote` / `409 already_voted`. |
| `POST` | `/api/rooms/{code}/host/unlock` | `{pin}` | Sets host cookie. `401 invalid_pin`, `429 pin_locked`. |

### 12.3 Host (host cookie required)

| Method | Path | Body | Notes |
| --- | --- | --- | --- |
| `POST` | `/api/rooms/{code}/playback` | `{action:"play"|"pause"|"next"|"cancelCountdown"}` | `next` advances with a notice; `cancelCountdown` holds. |
| `DELETE` | `/api/rooms/{code}/songs/{entryId}` | — | Remove pending entry. |
| `PUT` | `/api/rooms/{code}/order` | `{entryIds:[...]}` | Full permutation; sets `ManualOrder`. |
| `POST` | `/api/rooms/{code}/queue/clear` | `{}` | Removes all pending. |
| `POST` | `/api/rooms/{code}/queue/lock` | `{locked:true}` | — |
| `POST` | `/api/rooms/{code}/skip-vote/cancel` | `{}` | Clears active vote. |
| `POST` | `/api/rooms/{code}/pin/rotate` | `{}` | Returns new PIN (TV only in UI). |
| `POST` | `/api/rooms/{code}/end` | `{}` | Deletes room; broadcasts `roomEnded`. |

### 12.4 Snapshot format

The same shape is returned by `GET /api/rooms/{code}`, in WS `snapshot` messages, and after mutations (HTTP actions return the updated snapshot for convenience).

```json
{
  "seq": 42,
  "serverTime": "2026-10-21T20:10:00Z",
  "room": {
    "code": "blue-mic",
    "queueLocked": false,
    "manualOrder": false,
    "guestCount": 9,
    "guestCapacity": 25,
    "queueCapacity": 100
  },
  "nowPlaying": {
    "entryId": "q_abc",
    "videoId": "dQw4w9WgXcQ",
    "title": "Bohemian Rhapsody (Karaoke Version)",
    "channel": "Sing King",
    "thumbnailUrl": "https://i.ytimg.com/vi/.../mqdefault.jpg",
    "nickname": "Alex",
    "durationSeconds": 355,
    "status": "playing",
    "elapsedSeconds": 42.5,
    "generation": 7,
    "countdownEndsAt": null,
    "mine": false
  },
  "queue": [
    {
      "id": "q_def",
      "position": 1,
      "videoId": "...",
      "title": "...",
      "channel": "...",
      "thumbnailUrl": "...",
      "durationSeconds": 201,
      "nickname": "Sam",
      "score": 4,
      "votedByMe": true,
      "mine": false
    }
  ],
  "skipVote": null,
  "notices": [
    {"id":"n_1","code":"notice.songAdded","params":{"nickname":"Alex","title":"..."},"createdAt":"..."}
  ],
  "you": {
    "deviceId": "dev_...",
    "role": "guest",
    "hostUnlocked": false,
    "canSkipVote": true,
    "inQueue": true
  }
}
```

Guest snapshots must omit `hostToken`, `pin`, device cookie values, and any internal maps.

### 12.5 WebSocket

- URL: `/ws?room={code}`. Cookies authenticate; the server derives role (`tv` if host cookie + `client=tv`, `host_phone` if host cookie, else `guest`).
- The client opens with `client=guest|tv|phone`.
- Server → client:
  - `{type:"snapshot", seq, ...}` — full snapshot, coalesced at 100 ms.
  - `{type:"roomEnded"}` — followed by close code 4001.
  - `{type:"pong", serverTime}`.
- Client → server:
  - `{type:"heartbeat"}` every 20 s.
  - `{type:"tvProgress", entryId, generation, positionSeconds, durationSeconds}` every 5 s while playing.
  - `{type:"tvEnded", entryId, generation}`.
  - `{type:"tvError", entryId, generation, code}`.
- On connect: validate room/cookie, add socket, send one snapshot immediately. On auth failure: HTTP `403` before upgrade where possible.
- Ping/pong at the WS protocol level every 30 s; close idle sockets after 60 s without heartbeat.
- Backpressure: buffer 16 snapshots; if full, drop the socket; the client reconnects and resyncs.
- Snapshot `seq` increments per room on every state version; clients discard lower seq.
- Host heartbeat also updates `HostLastSeenAt`; guest heartbeats update device activity.
- No client mutation over WS in v1.

---

## 13. Security and abuse (public internet)

Threat model: anonymous internet users, party guests sharing a link, a host who may lose their TV link, and external Invidious dependencies. No sensitive data is stored, so the main risks are abuse, unauthorized room control, and XSS.

Controls:

- **Host token.** 128-bit random, base64url, constant-time comparison. Only in the `/tv/{token}` path; never in guest responses or logs. `Referrer-Policy: no-referrer` on all pages.
- **Host cookie.** HMAC-signed value containing room code, issue time, expiry (12 h). HttpOnly, Secure, SameSite=Lax, path `/`. Per-room cookie name avoids cross-room confusion. Host unlock rotates the session ID? Not necessary; PIN rotation invalidates only future unlocks.
- **PIN.** 4 digits, shown on TV by design. Rate limit 5 attempts per 10 minutes per IP+room, then 15-minute lockout. Log failures at `warn` without the PIN. Do not use the PIN as the only protection for the TV.
- **Device cookie.** Opaque random 128-bit ID plus HMAC; HttpOnly, SameSite=Lax, Secure in prod, 7-day expiry, renewed on use. Not a login.
- **CSRF.** SameSite=Lax cookies + require `Origin`/`Sec-Fetch-Site` to be same-origin for all non-GET requests. No CORS headers; browsers cannot read cross-origin responses.
- **XSS.** Svelte escapes all content; nicknames/titles are text nodes, never `{@html}`. No user-provided URLs are rendered as links except validated YouTube IDs; thumbnails come from a validated/allowlisted host. CSP below.
- **SSRF.** Paste parsing is allowlist-based and extracts an ID; the server never fetches arbitrary URLs. Invidious instances come only from configuration. oEmbed uses a fixed YouTube host. No redirects followed for user input.
- **CSP.**
  ```
  default-src 'self';
  script-src 'self' https://www.youtube.com;
  frame-src https://www.youtube.com https://www.youtube-nocookie.com;
  img-src 'self' data: https://*.ytimg.com https://*.ggpht.com;
  style-src 'self' 'unsafe-inline';
  connect-src 'self';
  font-src 'self';
  object-src 'none';
  base-uri 'none';
  frame-ancestors 'none';
  form-action 'self';
  ```
  Vite's production output must not need inline scripts. If thumbnails come from a configured Invidious host, proxy them through `/api/thumb/{id}` or add the host explicitly; proxying is preferred to avoid a broad CSP.
- **Headers.** `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, `Permissions-Policy: camera=(), microphone=(), geolocation=()` (QR scanning is done by the phone's camera app, not the page), HSTS at the proxy.
- **Rate limits (in-memory token buckets, per IP):**
  - Create room: 5/hour.
  - Search: 30/minute.
  - Add song: 10/minute per device + 30/minute per IP.
  - Upvote: 60/minute per device.
  - Skip vote start/cast: 10/minute per device.
  - Host unlock: 5/10 minutes per IP+room.
  - WS connect: 10/minute per IP.
  - Exceeding returns `429` with `Retry-After`.
- **Headers/proxy.** `TRUST_PROXY=true` only behind a proxy that overwrites `X-Forwarded-For`; otherwise IPs can be spoofed. Document this in README.
- **Limits as DoS protection.** Max rooms 1,000, 25 guests, 100 queue entries, 20 search results, 120-char queries, 24-char nicknames, 64 KB request bodies, 4 KB WS messages.
- **Privacy.** No accounts, no analytics, no cookies beyond device/host/language (language is localStorage). IPs live only in memory for rate limiting. Logs omit nicknames by default (room code + counts only). The privacy page states that searches are sent to configured Invidious instances and playback to YouTube.
- **Content.** No automated moderation; room codes are unlisted, not secret from guests. Report/abuse tooling is out of scope for v1.

---

## 14. Reliability and edge cases

| Situation | Required behavior |
| --- | --- |
| Unknown room code | Friendly 404 page in the current language; join field preserved; no distinction between “never existed” and “ended” in public API. |
| Room ended while open | Overlay/screen “This room has ended”; disable inputs; offer create new room; WS closed. |
| Room full | Join page shows capacity message and a retry button; WS rejected with `room_full`. |
| Queue full | Add disabled with “100 songs queued — try again later”; server enforces. |
| Queue locked | Add disabled with localized reason; paste box disabled too. |
| Duplicate song | Allowed; both entries vote separately. |
| Search proxy down | Cached/stale results if available + banner; otherwise error + paste box. |
| Paste metadata unavailable | Add with oEmbed title/thumbnail and unknown duration; show a “duration unknown” badge. |
| Live/Short/too long | Filtered from search; pasted URLs rejected after metadata check with a clear reason. |
| Video unplayable | Auto-skip, notice naming the title, continue queue. |
| Autoplay blocked | “Tap to resume” overlay; no skip, no state change. |
| TV refresh | Snapshot + seek to elapsed; resume; overlay if blocked. |
| TV offline briefly | Player continues locally; on reconnect, snapshot may correct drift >2 s or advance if the song ended elsewhere. |
| Two TVs | Both follow commands; each plays its own audio; echo accepted. |
| Host TV tab backgrounded | Browser may throttle timers; WS heartbeat still updates host presence. Playback may pause due to browser policy; TV shows overlay on return. |
| Host closes TV | Room persists for 2 h; host can reopen via the same secret link; phones retain queue. |
| Host phone leaves | Queue remains; TV is still host; room unaffected. |
| Skip vote ties | Threshold is strict majority; a tie fails. |
| Eligible voter disconnects | Their yes stays; threshold stays frozen. |
| Song ends during vote | Vote discarded; normal advance. |
| Host skips during vote | Vote cleared; direct skip; no vote notice. |
| Countdown cancel | Hold on next-song card; host Play starts it; no auto-start. |
| Empty queue after song | Idle screen with QR; no black player. |
| Server restart | All rooms lost; clients show a clear notice and create-room CTA. |
| Clock skew | Server timestamps only; clients compute offsets from `serverTime`; UI countdowns clamp at 0. |
| Long title/nickname | Truncate visually with ellipsis; full title in tooltip/aria-label. |
| French accents in codes | Room-code alphabet is ASCII words; normalization strips accents/diacritics from input. |
| Double-tap add/vote | Buttons disable in flight; server operations are idempotent per action where possible. Vote toggle is not idempotent by design; UI debounces 300 ms. |
| Back/forward navigation | SPA route handling; leaving TV does not end room. |
| Expired host cookie | Host bar disappears; PIN unlock again; TV still works via secret link. |

---

## 15. Observability

- `log/slog` JSON logs to stdout: request method/path/status/duration/bytes, room create/end (code, counts), search provider failures (instance host, status), WS connect/disconnect (role, room), rate-limit hits, cleanup deletions.
- Never log PINs, host tokens, cookies, or nicknames. Song/video IDs may be logged at debug level.
- Request ID middleware (`X-Request-ID` passthrough or generated) included in error logs.
- Counters exposed only in logs in v1: rooms, guests, search cache hit/miss, stale serves, failed playback skips.
- `/healthz` is liveness only; do not make it depend on Invidious.
- Startup logs configuration summary with secrets redacted, including whether Invidious instances are configured and the app version/commit.
- Graceful shutdown on SIGINT/SIGTERM with a 10 s timeout.

---

## 16. Testing strategy and acceptance criteria

### 16.1 Unit tests (domain / use cases, no network)

- Queue ordering: score descending, ties by insertion; new zero-vote entries append; manual order freezes vote sorting; reorder permutation validation.
- Vote toggle: one per device, unvote, score recomputation, duplicate video IDs independent.
- Skip vote: eligibility set (queue owners + current singer, 2-minute activity, frozen snapshot), threshold for 1/2/3/9/10 eligible, early pass, expiry, cooldown, cancel, song change clears vote.
- Playback state: elapsed math for play/pause/resume/seek, clamp, countdown deadline, generation guard against stale TV events, empty queue → idle.
- Room rules: guest cap, queue cap, lock, duplicates, nickname validation, room code normalization/collision retries, host token/PIN rotation.
- Lifecycle: 2 h no-host cleanup, 15 min never-opened cleanup, max rooms, end-room broadcast.
- Search normalization: suffix logic (don't double-append “karaoke”), filters (live, Shorts, >600 s), dedupe, URL parser positive/negative cases, SSRF rejection.
- i18n key parity test: every key in `en.ts` exists in `fr.ts` and vice versa; no hardcoded strings in components (lint rule or targeted scan).

### 16.2 Adapter tests

- HTTP with `httptest`: create/join/add/vote/skip/unlock/host actions; cookie flags; host required; capacity errors; `429`; content-type/origin checks; SPA fallback does not shadow `/api`.
- WebSocket with `coder/websocket` client: initial snapshot, broadcast on add, coalescing, room end close, heartbeat/presence, stale seq ignored, TV telemetry accepted/ignored by generation.
- Invidious client against a fake server: success, timeout, 500, malformed JSON, instance rotation, circuit breaker, stale cache, oEmbed fallback.
- QR endpoint returns valid SVG and encodes the configured public URL.
- Race detector on use case + hub tests (`go test -race ./...`).

### 16.3 Frontend tests

- Vitest + Testing Library: welcome create/join validation, language switch, search rendering/empty/error/stale, nickname sheet validation, add flow, queue sorting/vote toggle, skip-vote states, host unlock and host bar, countdown display from server deadline, reconnect banner, room-ended screen.
- Playwright, two browser contexts + one TV page:
  1. Create room → TV shows code/QR/PIN.
  2. Guest scans equivalent URL → search (mocked API) → add song with nickname → queue visible on TV.
  3. Second guest upvotes → order changes if applicable.
  4. Host pauses/resumes; guest sees state.
  5. Start skip vote → cast → pass → countdown → next song.
  6. Host drags/removes/locks/clears; guest UI reflects.
  7. Host unlocks phone with PIN; wrong PIN throttled.
  8. End room → all screens show ended.
  9. TV reload mid-song → position resumed within tolerance (use a fake player or a test video).
- Accessibility smoke: axe on welcome/guest/TV; keyboard-only host unlock and queue actions.

### 16.4 Manual acceptance script (v1 done)

1. `make dev` or `docker run` on a laptop; open TV view on a 1080p screen; QR scans to a phone on the same network.
2. Search a common karaoke song in both languages; results are relevant and filtered.
3. Add 5 songs from 3 phones; TV queue order matches; upvotes reorder; host drag freezes order.
4. Play through a song; auto-advance countdown appears; cancel and resume.
5. Skip a song by vote with a majority; verify failing vote and cooldown.
6. Trigger a known non-embeddable video; verify notice and auto-skip.
7. Refresh TV mid-song; verify resume; pull network; verify reconnect.
8. Lock queue, attempt add, clear queue, rotate PIN, end room.
9. Check no console errors; cookies are HttpOnly; host token absent from guest JSON; CSP does not break the player.
10. Run `go test -race ./...`, `go vet ./...`, `gofmt -l .`, `npm run check`, `npm run test`, `npm run build` — all clean.

### 16.5 Definition of done

- All v1 functional rules above implemented and covered by at least one automated test.
- No known security blocker from §13.
- README documents setup, env vars, deployment, and the Invidious dependency.
- The spec's contradictions are resolved in code as written here (vote ordering, manual order, skip-vote eligibility).
- Performance on a 2020-era phone: initial guest page < 2 s on LAN, search response < 1.5 s with a warm cache, snapshot broadcasts < 50 ms server time.

---

## 17. Build and deployment

### 17.1 Local development

```bash
# terminal 1 — Go API on :8080
go run ./cmd/server

# terminal 2 — Vite dev server on :5173, proxying /api and /ws
cd web && npm install && npm run dev
```

Vite proxy config:

```ts
server: {
  proxy: {
    '/api': 'http://localhost:8080',
    '/ws': { target: 'ws://localhost:8080', ws: true }
  }
}
```

`DEV=true` relaxes Secure cookies and allows `http://localhost`. The Go server serves the built SPA for `/`, `/r/*`, `/tv/*` when `web/dist` exists; in dev, use Vite.

### 17.2 Production build (single binary)

```bash
cd web && npm ci && npm run build     # outputs web/dist
cd .. && go build -o fana-karaoke ./cmd/server
./fana-karaoke
```

- `web/embed.go` contains `//go:embed all:dist` and exports the filesystem.
- The Go server serves fingerprinted assets with `Cache-Control: public, max-age=31536000, immutable`; `index.html` with `no-cache`.
- `make build` chains the two steps. `make test` runs Go + frontend tests. `make dev` runs both processes (or documents two terminals).

### 17.3 Container

Multi-stage Dockerfile:

1. `node:22-alpine`: `npm ci`, `npm run build`.
2. `golang:1.26-alpine`: copy source + `web/dist`, `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w"`.
3. `alpine:3.22` (or distroless): non-root user, copy one binary, expose `8080`, `HEALTHCHECK` on `/healthz`.
4. Runtime env: `PUBLIC_BASE_URL`, `INVIDIOUS_INSTANCES`, `SESSION_SECRET`; optional `MAX_ROOMS`, etc.
5. No database volume; nothing persists. Document that restarting loses rooms.

### 17.4 Reverse proxy requirements

- HTTPS termination and HSTS.
- Forward `Upgrade`/`Connection` for `/ws`, no response buffering/compression on WS, idle timeout > 60 s.
- Pass the real client IP (`X-Forwarded-For`) and set `TRUST_PROXY=true` only if the proxy overwrites it.
- Do not cache `/api/*` or WebSocket; cache only `/assets/*`.
- `PUBLIC_BASE_URL` must match the external URL so QR codes and YouTube `origin` are correct.
- Single instance only. Horizontal scaling would require shared room state and a shared pub/sub; out of scope.

---

## 18. Implementation plan

| Milestone | Scope | Exit criteria |
| --- | --- | --- |
| M0 — Skeleton | Go module, config, Svelte/Vite scaffold, embed pipeline, healthz, Dockerfile | `make build` produces one binary serving a Svelte welcome page. |
| M1 — Rooms | Domain Room/QueueEntry, memory repo, code generation, create/join, host token, PIN, cookies | Create room via API; TV and guest shells load correct room; 404/full paths tested. |
| M2 — Realtime | WS hub, snapshots, seq, reconnect, connection dots | Two clients see the same room; reconnect resyncs; race tests pass. |
| M3 — Search | Invidious client, rotation, cache/stale, oEmbed, paste parser, search + add metadata | Search and paste add songs with correct filters; fake-server tests pass. |
| M4 — Queue | Add with nickname, list, upvotes, vote sorting, duplicate/limits, queue lock | Queue rules covered by unit + HTTP tests; UI reflects order. |
| M5 — Playback | YouTube IFrame player, start gesture, pause/resume/next, countdown, resume-at-position, error skip | TV plays, auto-advances, resumes after refresh; fake-player tests cover state machine. |
| M6 — Skip votes | Eligibility, threshold, 60 s window, cooldown, host bypass/cancel | Majority pass and failure paths tested end-to-end. |
| M7 — Host admin | Phone PIN unlock, host bar, remove/reorder/clear/lock/rotate/end, confirmations | Both host surfaces work; authorization tests pass. |
| M8 — Polish | EN/FR copy, neon design, a11y, mobile layout, QR, privacy/help | Language parity, axe smoke, responsive screenshots; no untranslated strings. |
| M9 — Hardening | Rate limits, CSP/headers, cleanup, logging, Docker, README, Playwright suite | Definition of done in §16.5 met. |

---

## 19. Open questions and risks

These do not block starting; decide during implementation or with the user.

1. **Invidious availability.** Public instances are frequently rate-limited or offline, and their terms/usage limits vary. The cache/stale/paste/oEmbed fallbacks mitigate this, but search quality is not guaranteed. A future option is a self-hosted instance or a YouTube Data API key.
2. **Word list.** Pick a permissively licensed, curated two-syllable English word list (~1,000–2,000 words) and store it in the Go binary. Confirm the license and avoid ambiguous/offensive words. Two words yield ~1–4 million combinations; collision retries handle the rest.
3. **Shorts detection.** Duration ≤ 60 s is a heuristic; YouTube Shorts can be longer. If this proves noisy, filter by the Invidious result URL/type or accept a small number of Shorts.
4. **TV-browser autoplay.** Some TV browsers block programmatic playback even after a gesture. The “Tap to resume” overlay is the fallback; the host may need a Bluetooth mouse/keyboard or a laptop connected to the TV. Document this.
5. **Hidden YouTube controls.** `controls=0` may hide captions and settings. Karaoke videos have burned-in lyrics, so this is acceptable; verify the player still shows an error/consent state and that YouTube branding requirements are met.
6. **PIN is not security.** Anyone who can see the TV can become host. This is intentional for a party, but it means a public room's host control is only as private as the room. The TV secret link protects remote takeover; if stronger host auth is needed, add a host password at creation (future).
7. **French copy and room words.** English room-code words in a French UI are accepted, but the word list should avoid words that are awkward in French.
8. **Search region/language.** Default region FR; some karaoke catalogs are region-sensitive. Make `YOUTUBE_REGION` configurable and revisit after testing.
9. **Queue fairness.** Votes can let one table dominate. Host manual ordering and the 100-song cap are the current mitigations; a round-robin mode is out of scope.
10. **Multiple TVs echo.** Accepted. If it becomes a real problem, add a “mute this TV” local toggle (not a server feature).
11. **Persistent rooms.** In-memory means a restart clears the party. Accepted; document clearly. If needed later, an optional SQLite adapter can be added behind `RoomRepository` without touching the domain.
12. **No lyrics overlay.** YouTube karaoke videos include lyrics; if a user adds an original track, there are no lyrics. The UI can warn in the add sheet (copy TBD) but does not analyze content.
13. **HTTPS on LAN.** Phone camera QR scanning and secure cookies generally require HTTPS or localhost. For a LAN party on `http://192.168.x.x`, Secure cookies would break; `DEV=true` or a local TLS certificate is required. Document this deployment caveat. Public deployment assumes HTTPS.

---

## 20. Appendix — glossary

- **Room:** one karaoke session, identified by a readable code, with one queue and one now-playing item.
- **TV view / playback screen:** the host-authenticated page with the YouTube player and QR code.
- **Guest view:** the phone page for search, queue, votes, and (optionally) host controls.
- **Host token:** the secret in `/tv/{token}` that grants host authority without a PIN.
- **PIN:** the 4-digit code shown on the TV, used to unlock host controls on a phone.
- **Queue:** pending songs, sorted by votes unless the host has taken manual order.
- **Now playing:** the song currently loaded on the TVs; no longer in the queue.
- **Skip vote:** a 60-second, majority vote among eligible queue owners to end the current song.
- **Upvote:** a per-device toggle that raises a queued song's score and (before manual order) its position.
- **Countdown:** the 5-second gap between songs, cancellable by the host.
- **Notice:** an ephemeral localized event shown as a toast/ticker and included in snapshots.
