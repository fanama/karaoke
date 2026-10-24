.PHONY: build web server dev-web test fmt e2e clean worker-install worker-test worker-dev worker-deploy

build: web
	go build -o fana-karaoke ./cmd/server

web:
	cd web && npm ci && npm run build

server:
	go run ./cmd/server

dev-web:
	cd web && npm run dev

test:
	go test ./... -race
	go vet ./...
	cd web && npm run check && npm run build

e2e:
	cd web && node e2e/flow.mjs && node e2e/autoplay.mjs && node e2e/audit.mjs && node e2e/voting.mjs

worker-install:
	cd workers/library && npm ci

worker-test:
	cd workers/library && npm test && npm run check

worker-dev:
	cd workers/library && npx wrangler dev --local

worker-deploy:
	cd workers/library && npx wrangler deploy

fmt:
	gofmt -w cmd internal web/embed.go

clean:
	rm -f fana-karaoke
	rm -rf web/dist/assets
