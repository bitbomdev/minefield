# Default target
.DEFAULT_GOAL := all

# Build target
build:
	go build -o bin/bitbom main.go

# Test target
test:
	go test -v ./...

# Clean target
clean:
	rm -rf bin

# Clean Redis data
clean-redis:
	docker-compose exec redis redis-cli FLUSHALL

# Docker targets
docker-up:
	docker-compose up -d

docker-down: clean-redis
	docker-compose down

docker-logs:
	docker-compose logs -f

all: build test # Build and test the project
