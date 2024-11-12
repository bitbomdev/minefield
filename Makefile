# Default target
.DEFAULT_GOAL := all

# Build target
build:
	go build -o bin/minefield main.go

# Test target
test:
	go test -v -coverprofile=coverage.out ./...

test-e2e: docker-up
	e2e=true go test -v -coverprofile=coverage.out ./...

# Clean target
clean:
	rm -rf bin

# Clean Redis data
clean-redis:
	docker compose exec -T redis redis-cli ping || docker compose up -d redis
	docker compose exec -T redis redis-cli FLUSHALL

# Docker targets
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

all: build test docker-build go-mod-tidy git-porcelain

.PHONY: test test-e2e build clean clean-redis docker-up docker-down docker-logs docker-build all buf-generate install-buf
