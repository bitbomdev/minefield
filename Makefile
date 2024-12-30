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

define print_coverage
	@echo "==> Coverage Report"
	@echo "==> Per Package Coverage (excluding generated and mock files):"
	@go tool cover -func=coverage.out | \
		grep -v "gen/" | \
		grep -v "mockGraph.go" | \
		grep -v "total:" | \
		awk '{print $$3 " " $$1 " " $$2}' | \
		sed 's/%//' | \
		sort -nr | \
		awk '{print $$2 " " $$3 " " $$1 "%"}'
	@echo "\n==> Total Coverage (excluding generated and mock files):"
	@go tool cover -func=coverage.out | \
		grep -v "gen/" | \
		grep -v "mockGraph.go" | \
		grep "total:" | \
		awk '{print "Total coverage: " $$3}'
	@echo "\n==> HTML Coverage Report:"
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"
endef

coverage: test test-e2e
	$(call print_coverage)

.PHONY: test test-e2e build clean clean-redis docker-up docker-down docker-logs docker-build all wire coverage