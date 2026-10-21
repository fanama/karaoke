# Fana Karaoke

A browser karaoke room. Put the room on the TV, let guests scan a QR code, search YouTube for karaoke videos, upvote the queue, and vote to skip. No accounts, no installs, no database.

- **TV view** (`/tv/<secret>`) — the YouTube player, join QR, host PIN, queue drawer, and host controls.
- **Guest view** (`/r/<code>`) — phone-first search, paste-a-link, queue, upvotes, skip votes, and host controls after entering the PIN.
- Rooms are **ephemeral**: in memory, gone when the host ends them, after two hours without a host, or on restart.

The product and design rationale live in [`fana-karaoke-spec.md`](fana-karaoke-spec.md).

## Stack

- Go 1.25 HTTP server, in-memory room manager, WebSocket snapshots, signed cookies.
- Svelte 5 + TypeScript + Vite, built into `web/dist` and embedded with `go:embed`.
- YouTube IFrame Player API for playback; no media is downloaded or proxied.
- Public [Invidious](https://invidious.io/) instances for search and metadata, with a 24-hour cache, stale results, and YouTube oEmbed fallback.
- No database, no frontend runtime server, no external font/CDN dependency at runtime.

## Run locally

Requirements: Go 1.25+, Node 22+.

```bash
# Build the frontend once (embedded into the binary)
cd web
npm ci
npm run build
cd ..

# Run
go run ./cmd/server
# open http://localhost:8080
```

Or build one binary:

```bash
make build
./fana-karaoke
```

### Frontend development

Two terminals:

```bash
# terminal 1
DEV=true go run ./cmd/server

# terminal 2
cd web
npm run dev
# open http://localhost:5173
```

Vite proxies `/api` and `/ws` to port 8080. `DEV=true` disables rate limits and relaxes the WebSocket origin check for the Vite origin.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | HTTP listen port. |
| `PUBLIC_BASE_URL` | derived from request | Absolute origin for QR codes, join/TV links, and secure cookies. Set it in production, e.g. `https://karaoke.example`. |
| `INVIDIOUS_INSTANCES` | `https://invidious.f5.si,https://yewtu.be,https://inv.nadeko.net,https://invidious.nerdvpn.de,https://iv.melmac.space` | Comma-separated Invidious base URLs, tried in order. Unhealthy instances are skipped for 60 s. |
| `YOUTUBE_REGION` | `FR` | Region passed to Invidious search. |
| `SESSION_SECRET` | random at boot | HMAC key for device/host cookies. Set and keep stable in production. |
| `ROOM_TTL` | `2h` | Delete a room with no host connection for this long. |
| `MAX_ROOMS` | `1000` | Server-wide room cap. |
| `MAX_GUESTS` | `25` | Concurrent guest devices per room. |
| `MAX_QUEUE` | `100` | Pending songs per room. |
| `SEARCH_CACHE_TTL` | `24h` | Fresh search-cache lifetime. |
| `SEARCH_STALE_TTL` | `168h` | Maximum age of a cached result served while Invidious is down. |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error`. |
| `TRUST_PROXY` | `false` | Use `X-Forwarded-For` for rate limits. Only enable behind a proxy that overwrites it. |
| `DEV` | `false` | Relax Secure cookies/rate limits and skip WebSocket origin checks. Never enable in production. |

Example:

```bash
PUBLIC_BASE_URL=https://karaoke.example \
SESSION_SECRET="$(openssl rand -hex 32)" \
INVIDIOUS_INSTANCES="https://invidious.f5.si" \
./fana-karaoke
```

## Docker

```bash
docker build -t fana-karaoke .
docker run --rm -p 8080:8080 \
  -e PUBLIC_BASE_URL=http://localhost:8080 \
  -e SESSION_SECRET="$(openssl rand -hex 32)" \
  fana-karaoke
```

The image is a multi-stage build: Node builds the Svelte assets, Go produces a static binary, and the runtime image is Alpine with CA certificates. No volume is needed because there is no persistent state.

## Public deployment notes

- Terminate TLS at a reverse proxy and set `PUBLIC_BASE_URL` to the HTTPS origin. Phone QR scanning and Secure cookies expect HTTPS.
- Forward WebSocket upgrades on `/ws` (`Upgrade`/`Connection`), disable response buffering/compression for that path, and keep idle timeouts above 60 s.
- Run **one instance**. Room state and the WebSocket hub are process-local; horizontal scaling would need shared state and pub/sub.
- Set `SESSION_SECRET`. If it changes, existing host/device cookies are invalidated (the rooms themselves are lost on restart anyway).
- Set `TRUST_PROXY=true` only when the proxy overwrites `X-Forwarded-For`; otherwise rate limits can be spoofed.
- `/healthz` is liveness-only and does not depend on Invidious.

## Search behavior

- Queries are normalized and ` karaoke` is appended unless the query already contains it.
- Live streams, videos longer than 10 minutes, videos under 60 seconds, and videos whose owners disabled embedding are filtered out.
- Results are cached for 24 hours and stale results can be served for up to a week with a banner.
- Pasting a YouTube URL always works. Metadata comes from Invidious first, then YouTube oEmbed (duration may be unknown until the player reports it).
- Public Invidious instances are third-party services and can fail or rate-limit. Point `INVIDIOUS_INSTANCES` at a reliable instance (or a self-hosted one) for production. No YouTube Data API key is required.

## Rules in brief

- Host access: secret `/tv/<token>` link + a random 4-digit PIN for the phone.
- Queue: upvote-sorted; new songs start at the bottom; once the host reorders, **host order wins** for the rest of the room.
- Upvotes: one toggle per device per song; the same video can be queued multiple times.
- Skip vote: available to people who own a queued song or are currently singing. 60 seconds, majority of the frozen eligible set, 30-second cooldown after failure. Hosts can skip directly.
- Auto-advance: 5-second countdown, cancellable by the host.
- Playback failure: auto-skip with a room notice.
- Refresh the TV: it reloads the current video and seeks to the server-known position.
- Multiple TVs: each plays independently and follows the same host commands.

## Architecture

```text
cmd/server                 composition root / config / graceful shutdown
internal/domain            entities, views, error codes
internal/app               room manager: queue, votes, playback, cleanup
internal/search            Invidious client, cache, oEmbed fallback
internal/video             safe YouTube URL parser
internal/httpapi           routes, cookies, host auth, WS hub, QR
web/                       Svelte 5 app, embedded assets, Playwright e2e
```

The room manager is independent of HTTP and Svelte. The HTTP adapter owns cookies, rate limits, and the WebSocket hub. Mutations go through HTTP; WebSockets carry full snapshots plus TV playback telemetry.

## Verify

```bash
go test ./... -race
go vet ./...
gofmt -l .
cd web && npm run check && npm run build
```

Browser flows (requires Chrome installed and the server running on port 8090):

```bash
# terminal 1
PORT=8090 DEV=true PUBLIC_BASE_URL=http://localhost:8090 go run ./cmd/server

# terminal 2
cd web
node e2e/flow.mjs
node e2e/autoplay.mjs
node e2e/audit.mjs
node e2e/voting.mjs
```

The scripts exercise create/join, QR, paste-add, host unlock, live queue updates, countdown, drawer, end room, autoplay with and without a user gesture, French locale, responsive overflow, search results, and a majority skip vote.

## Limitations

- Rooms and host sessions are lost on restart. This is intentional.
- The public Invidious dependency is best-effort; search can be unavailable while URL paste keeps working.
- The 4-digit PIN is a convenience, not authentication — it is displayed on the TV. The secret TV token is what protects remote host access.
- YouTube playback requires the official iframe player; the video is never hidden or downloaded.
- Autoplay: the TV tries to start each song with sound. If the browser refuses (no prior interaction, strict TV browser), playback starts muted instead of stalling and a **Tap for sound** chip appears — one tap anywhere on the TV restores audio. A hard block still shows **Tap to resume**.
- Not implemented: lyrics overlay, scoring, vocal removal, saved playlists, accounts, moderation, native apps.
