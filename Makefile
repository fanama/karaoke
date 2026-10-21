.PHONY: build web server dev-web test fmt e2e clean

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

fmt:
	gofmt -w cmd internal web/embed.go

clean:
	rm -f fana-karaoke
	rm -rf web/dist/assets
