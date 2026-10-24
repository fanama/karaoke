# Fana Karaoke

A browser karaoke room. Put the room on the TV, let guests scan a QR code, search YouTube for karaoke videos, upvote the queue, and vote to skip. No accounts, no installs. A small SQLite library remembers songs so repeat searches skip YouTube.

- **TV view** (`/tv/<secret>`) — the YouTube player, join QR, host PIN, queue drawer, and host controls.
- **Guest view** (`/r/<code>`) — phone-first search, paste-a-link, queue, upvotes, skip votes, and host controls after entering the PIN.
- Rooms are **ephemeral**: in memory, gone when the host ends them, after two hours without a host, or on restart.

The product and design rationale live in [`fana-karaoke-spec.md`](fana-karaoke-spec.md).

Created by [Fanamperantsoa Rakotoasimbola](https://github.com/fanama).

## Stack

- Go 1.25 HTTP server, in-memory room manager, WebSocket snapshots, signed cookies.
- Svelte 5 + TypeScript + Vite, built into `web/dist` and embedded with `go:embed`.
- YouTube IFrame Player API for playback; no media is downloaded or proxied.
- Public [Invidious](https://invidious.io/) instances for search and metadata, with a 24-hour cache, stale results, and YouTube oEmbed fallback.
- A local SQLite song library (pure-Go [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite), no CGO) searched before Invidious. It keeps one row per video, remembers online hits and played songs, and can be exported/imported by the host.
- No frontend runtime server, no external font/CDN dependency at runtime. The song library is the only persistent file.

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
| `LIBRARY_PATH` | `./data/library.sqlite` | SQLite song library file. Created on boot; an empty value disables the library. Point it at a persistent volume in production. |
| `ADMIN_TOKEN` | unset | Enables the `/admin` page. Set a long random value, for example `openssl rand -hex 32`. Leave unset to disable the page. |
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

For local development, `cmd/server` also loads `.env` from the working directory at startup when the file exists. Variables already set in the real environment always win, so production settings are never overridden. The Docker image does not contain your `.env`; pass variables with `--env-file .env` or the platform's environment settings.

## Docker

```bash
docker build -t fana-karaoke .
docker run --rm -p 8080:8080 -v fana-data:/data \
  -e PUBLIC_BASE_URL=http://localhost:8080 \
  -e SESSION_SECRET="$(openssl rand -hex 32)" \
  fana-karaoke
```

The image is a multi-stage build: Node builds the Svelte assets, Go produces a static binary, and the runtime image is Alpine with CA certificates. Rooms still have no persistent state, but the song library lives at `/data/library.sqlite` (`LIBRARY_PATH` is preset in the image); mount a volume to keep it across container replacements.

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
- Attach a **Render disk** (paid plans) mounted at `/data` so the song library survives deploys. Without it the library still works but resets with the instance.

## Public deployment notes

- Terminate TLS at a reverse proxy and set `PUBLIC_BASE_URL` to the HTTPS origin. Phone QR scanning and Secure cookies expect HTTPS.
- Forward WebSocket upgrades on `/ws` (`Upgrade`/`Connection`), disable response buffering/compression for that path, and keep idle timeouts above 60 s.
- Run **one instance**. Room state and the WebSocket hub are process-local; horizontal scaling would need shared state and pub/sub.
- Set `SESSION_SECRET`. If it changes, existing host/device cookies are invalidated (the rooms themselves are lost on restart anyway).
- Set `TRUST_PROXY=true` only when the proxy overwrites `X-Forwarded-For`; otherwise rate limits can be spoofed.
- Mount a persistent volume for `LIBRARY_PATH` (or leave it unset to disable the library).
- `/healthz` is liveness-only and does not depend on Invidious or the library.

## Search behavior

- The local SQLite library is searched first, matching every query word against the title and channel (accent-insensitive). A hit returns instantly and never calls Invidious.
- Online hits and added songs (including pasted links) are written to the library, deduplicated by YouTube id. Repeated queries therefore stop spending the upstream quota.
- Hosts can **Export** the library as a `.sqlite` file and **Import** one back from **Manage**. Import merges: known videos are refreshed, never duplicated.
- Queries are normalized and ` karaoke` is appended unless the query already contains it.
- Live streams, videos longer than 10 minutes, videos under 60 seconds, and videos whose owners disabled embedding are filtered out.
- Results are cached for 24 hours and stale results can be served for up to a week with a banner.
- Pasting a YouTube URL always works. Metadata comes from Invidious first, then YouTube oEmbed (duration may be unknown until the player reports it).
- Public Invidious instances are third-party services and can fail or rate-limit. Point `INVIDIOUS_INSTANCES` at a reliable instance (or a self-hosted one) for production. No YouTube Data API key is required.

## Admin page

Set a token and open `/admin`:

```bash
ADMIN_TOKEN="$(openssl rand -hex 32)" \
LIBRARY_PATH=./data/library.sqlite \
./fana-karaoke
```

Or put `ADMIN_TOKEN=…` in a local `.env` file and restart the server; `.env` is loaded automatically. `.env` is ignored by Git.

The page asks for the token once and keeps a signed, HttpOnly cookie for 24 hours. It shows the number of saved songs, downloads the SQLite file (**Download .sqlite**) and merges an uploaded file (**Upload .sqlite**) without creating duplicates. The same transfer is available to room hosts from the **Manage** panel.

Without `ADMIN_TOKEN`, `/admin` explains that the page is disabled and the admin API answers `404`. The admin cookie is signed with `SESSION_SECRET`; rotating that secret logs admins out.

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

```mermaid
flowchart TB
  subgraph Client["Browser — Svelte 5"]
    direction LR
    UI["ui<br/>welcome · guest · TV · admin"] --> App["application<br/>controllers + stores"]
    App --> FPorts["ports<br/>API · socket · player"]
    FAdapters["adapters<br/>fetch · WebSocket · YouTube"] -. implements .-> FPorts
  end

  subgraph Server["Go server"]
    direction LR
    HTTP["httpapi<br/>routes · cookies · rate limits · WS hub"]
    InPort["ports.KaraokeService"]
    Facade["usecase.Facade<br/>room · queue · playback · votes · library"]
    Domain["domain<br/>pure rules and entities"]
    OutPorts["ports output<br/>rooms · catalog · library · events · clock"]
    HTTP --> InPort
    InPort -. implemented by .-> Facade
    Facade --> Domain
    Facade --> OutPorts
  end

  subgraph Driven["Driven adapters"]
    direction LR
    Memory["memory<br/>rooms · devices · rate limits"]
    Invidious["invidious<br/>search · metadata · cache"]
    SQLLib["sqlite<br/>song library"]
    Support["bus · crypto · qrgen · clock"]
  end

  FAdapters -->|"HTTP /api + WebSocket /ws"| HTTP
  OutPorts -. implemented by .-> Driven
  Invidious --> External["Invidious instances / YouTube oEmbed"]
  SQLLib --> File[("data/library.sqlite")]
```

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
                                  (client, parser, oEmbed prober, cache), sqlite (song
                                  library), crypto (IDs, HMAC signer), qrgen, clock, bus
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

- Rooms and host sessions are lost on restart. This is intentional. Only the song library persists.
- The library stores song metadata (id, title, channel, thumbnail, duration), not audio or video.
- The public Invidious dependency is best-effort; search can be unavailable while URL paste keeps working.
- The 4-digit PIN is a convenience, not authentication — it is displayed on the TV. The secret TV token is what protects remote host access.
- YouTube playback requires the official iframe player; the video is never hidden or downloaded.
- Autoplay: the TV tries to start each song with sound. If the browser refuses (no prior interaction, strict TV browser), playback starts muted instead of stalling and a **Tap for sound** chip appears — one tap anywhere on the TV restores audio. A hard block still shows **Tap to resume**.
- Not implemented: lyrics overlay, scoring, vocal removal, saved playlists, accounts, moderation, native apps.
