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
| `PUBLIC_BASE_URL` | derived from request; on Render, `RENDER_EXTERNAL_URL` | Absolute origin for QR codes, join/TV links, secure cookies, and origin checks. Set it in production, e.g. `https://karaoke.example`. |
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

## Deploy on Render

The repository ships a [`render.yaml`](render.yaml) Blueprint:

1. Push the repository to GitHub and open **New → Blueprint** in the Render dashboard.
2. Select the repo; Render detects `render.yaml` and creates the `fana-karaoke` web service.
3. Apply. The first build takes a few minutes; the health check is `/healthz`.

The Blueprint builds the existing Dockerfile, keeps `TRUST_PROXY=true` (Render's proxy overwrites `X-Forwarded-For`), and generates a stable `SESSION_SECRET`. `PUBLIC_BASE_URL` falls back to Render's `RENDER_EXTERNAL_URL`, so QR codes, absolute links, Secure cookies and WebSocket origin checks work without extra configuration.

To deploy manually instead: **New → Web Service**, runtime **Docker**, health check path `/healthz`, and the same two environment variables.

**Constraints**

- Run **exactly one instance**. Room state and the WebSocket hub are process-local; scaling out breaks rooms and live updates.
- The free plan spins down after 15 minutes without traffic, which ends all rooms and adds a cold-start delay. Use a paid instance for an event.
- With a custom domain, set `PUBLIC_BASE_URL` explicitly (for example `https://karaoke.example`) so links and cookies use that origin.
- Deploys restart the process: rooms are lost and clients reconnect automatically.

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

Hexagonal architecture (ports & adapters) with a functional domain core. Dependencies point inward: the domain knows nothing about HTTP, storage, or Svelte; use cases only know the domain and the port interfaces.

```text
cmd/server/                       composition root: main.go, config.go, wire.go
internal/domain/                  pure core: entities + rules (room, queue, playback,
                                  voting, video, notice, snapshot/views, errors)
internal/ports/                   input.go (KaraokeService facade), output.go (repositories,
                                  event publisher, catalog, clock, IDs, signer, QR, limiter)
internal/usecase/                 room, queue, playback, vote and lifecycle services
                                  behind a single Facade
internal/adapters/primary/httpapi transport: routes, handlers, middleware, session cookies,
                                  SPA, WebSocket hub and handler
internal/adapters/secondary/      memory (room/device repos, rate limiter), invidious
                                  (client, parser, oEmbed prober, cache), crypto (IDs, HMAC
                                  signer), qrgen, clock, bus
internal/architecture/            automated layer and size checks
```

```text
web/src/domain/                   pure TypeScript: models, playback math, room code,
                                  autoplay policy
web/src/ports/                    contracts: API, socket, player, browser, DI context
web/src/application/              controllers and stores (welcome, guest, tv, i18n);
                                  no direct fetch, WebSocket or window.YT
web/src/adapters/                 fetch API, WebSocket, YouTube IFrame player, browser APIs
web/src/ui/                       Svelte components: common, welcome, guest, tv
web/src/styles/                   tokens, base, components
```

**Backend.** `internal/domain` is a set of pure functions over copied room state: no I/O, no mutexes, no transport types. `internal/usecase` orchestrates them through the driven ports (rooms, devices, clock, IDs, catalog, event publisher) in a `usecase.Deps` bundle. `internal/adapters/primary/httpapi` implements the driving side: it depends on the `ports.KaraokeService` interface, owns cookies, origin/CSRF checks, rate limits, QR responses and the WebSocket hub. `internal/adapters/secondary/*` provide the concrete driven adapters. `cmd/server/wire.go` is the only production code that constructs concrete types and wires them together; the event `bus` is connected to the hub there instead of via post-construction setters.

**Frontend.** `main.ts` builds the production adapters once and `App.svelte` exposes them through a Svelte context (`web/src/ports/context.ts`). Views ask `getAppContext()` for their dependencies and instantiate an application controller; UI components never import `adapters`, and the controllers never touch `fetch`, `WebSocket` or `window.YT` directly.

**Data flow.** Mutations go through HTTP; the use case persists the new room state and publishes room-updated / room-ended events on the bus, which the hub turns into full `snapshot` messages (and a final `roomEnded` before closing) over WebSockets. The TV sends playback telemetry (`tvEnded`, `tvError`, `tvProgress`) back over the same socket.

Layer isolation and the 300-line file / 40-line function limits are enforced by `go test ./internal/architecture/...`. The full rationale and dependency matrix are in [`ddd-architecture-spec.md`](ddd-architecture-spec.md).

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
