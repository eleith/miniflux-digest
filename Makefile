build:
	go build -mod=vendor -o miniflux-digest .

test:
	go test -mod=vendor ./...

test-coverage:
	go test -mod=vendor -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

lint:
	golangci-lint run

vendor:
	go mod vendor

ci: vendor lint test-coverage

all: ci build
