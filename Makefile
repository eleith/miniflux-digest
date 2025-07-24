build:
	go build -mod=vendor -o miniflux-digest ./cmd/miniflux-digest

preview-html:
	go run -mod=vendor ./scripts/preview-html/main.go
	./scripts/open-preview.sh

preview-miniflux-email:
	go run -mod=vendor ./scripts/preview-miniflux-email/main.go

test:
	go test -mod=vendor ./... ./cmd/miniflux-digest

test-coverage:
	./scripts/check-coverage.sh 60

lint:
	golangci-lint run

vendor:
	go mod vendor

all: ci build
