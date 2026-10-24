# Fana Library Worker

Small authenticated proxy between the Render Go service and a free Cloudflare D1
database. It exposes only:

- `GET /v1/health`
- `GET /v1/songs?limit=500&offset=0`
- `POST /v1/songs/upsert` (max 100 songs)
- `POST /v1/songs/import` (max 2,000 songs, one atomic D1 batch)

Every request needs `Authorization: Bearer <LIBRARY_PROXY_TOKEN>`. This is a
shared secret for this Worker, not a Cloudflare API token. No raw SQL is
accepted from Render.

## First deployment

```bash
cd workers/library
npm install
npx wrangler login              # opens the browser; no API token to manage
npx wrangler d1 create fana-library --location=weur
```

Copy the printed database id into `wrangler.toml` (replace
`REPLACE_WITH_DATABASE_ID`). Then:

```bash
npx wrangler d1 migrations apply fana-library --remote
openssl rand -hex 32            # copy this once
npx wrangler secret put LIBRARY_PROXY_TOKEN
npx wrangler deploy
```

The deploy prints a URL like
`https://fana-library.<account-subdomain>.workers.dev`.

Verify:

```bash
export FANA_LIBRARY_URL="https://fana-library.<account-subdomain>.workers.dev"
export FANA_LIBRARY_TOKEN="<the generated secret>"
curl -fsS -H "Authorization: Bearer $FANA_LIBRARY_TOKEN" \
  "$FANA_LIBRARY_URL/v1/health"
# {"ok":true}
```

Then set on Render:

- `LIBRARY_BACKEND=d1`
- `LIBRARY_PROXY_URL=<Worker URL>`
- `LIBRARY_PROXY_TOKEN=<same secret>`

## Local development

```bash
cd workers/library
npm install
npx wrangler d1 migrations apply fana-library --local
npx wrangler dev --local
```

Copy `.dev.vars.example` to `workers/library/.dev.vars` (ignored by Git). In
another terminal:

```bash
LIBRARY_BACKEND=d1 \
LIBRARY_PROXY_URL=http://127.0.0.1:8787 \
LIBRARY_PROXY_TOKEN=dev-secret \
go run ./cmd/server
```

Plain HTTP is accepted only for `localhost` / `127.0.0.1`.

## Tests

```bash
npm test
npm run check
```

The endpoint tests use a fake D1 and do not need a Cloudflare account. For a
real D1 integration check, use `npx wrangler dev --local` and `curl`.

## CI (optional)

For a CI deployment, create a scoped Cloudflare API token with **Workers
Scripts: Edit** and **D1: Edit** and store it as the GitHub Actions secret
`CLOUDFLARE_API_TOKEN`. Then run:

```bash
npx wrangler d1 migrations apply fana-library --remote
npx wrangler deploy
```

The token exists only in CI. It is never stored in Render.

Expose the token as a Worker secret, not a plain text variable:

```bash
npx wrangler secret put LIBRARY_PROXY_TOKEN
```

## Limits

- Workers Free: 100,000 requests/day, 10 ms CPU per invocation.
- D1 Free: 5 million rows read/day, 100,000 rows written/day, 500 MB per
  database. Limits reset at 00:00 UTC; data is not lost when a limit is hit.
- The cache load is 500 rows per page. A 2,000-song library costs about four
  Worker requests per Render wake.
- The import endpoint is one `batch()` call. Keep the 2,000-song cap so the
  batch stays atomic and within the free CPU budget.
