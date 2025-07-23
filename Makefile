build:
	go build -mod=vendor -o miniflux-digest .

preview:
	go run -mod=vendor ./scripts/preview/main.go
	./scripts/open-preview.sh

test:
	go test -mod=vendor ./...

test-coverage:
	go test -mod=vendor -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

lint:
	golangci-lint run

vendor:
	go mod vendor

all: ci build
