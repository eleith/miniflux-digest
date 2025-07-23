build:
	go build -mod=vendor -o miniflux-digest .

preview-html:
	go run -mod=vendor ./scripts/preview-html/main.go
	./scripts/open-preview.sh

preview-miniflux-email:
	go run -mod=vendor ./scripts/preview-miniflux-email/main.go

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
