.DEFAULT_GOAL := all

build: wire
	CGO_ENABLED=1 go build -o bin/minefield main.go

test:
	go test -v -coverprofile=coverage.out ./...

test-e2e: docker-up
	e2e=true go test -v -coverprofile=coverage.out ./...

clean:
	rm -rf bin

clean-redis:
	docker compose exec -T redis redis-cli ping || docker compose up -d redis
	docker compose exec -T redis redis-cli FLUSHALL

docker-up: docker-down
	docker compose up -d

docker-down: clean-redis
	docker compose down

docker-logs:
	docker compose logs -f

docker-build:
	docker build -t ghcr.io/bitbomdev/minefield:latest .

go-mod-tidy:
	go mod tidy

git-porcelain: 
	git status --porcelain

check-wire:
	@command -v wire >/dev/null 2>&1 || { echo >&2 "wire is not installed. Please install wire. go install github.com/google/wire/cmd/wire@latest"; exit 1; }

wire: check-wire
	cd cmd/server && wire || { echo "Wire generation failed in cmd/server"; exit 1; }

all: wire build test docker-build go-mod-tidy git-porcelain

.PHONY: test test-e2e build clean clean-redis docker-up docker-down docker-logs docker-build all wire