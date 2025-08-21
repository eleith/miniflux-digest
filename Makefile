build:
	go build -mod=vendor -o miniflux-digest ./cmd/miniflux-digest

preview:
	@trap "exit 0" INT; go run -mod=vendor ./scripts/preview/main.go

preview-html:
	go run -mod=vendor ./scripts/preview/main.go --html

preview-email:
	go run -mod=vendor ./scripts/preview/main.go --email

preview-miniflux:
	go run -mod=vendor ./scripts/preview/main.go --miniflux

test:
	go test -mod=vendor ./... ./cmd/miniflux-digest

test-coverage:
	./scripts/check-coverage.sh 60

test-coverage-full:
	./scripts/check-coverage.sh --mode=functions 60

lint:
	golangci-lint run

vendor:
	go mod vendor

all: ci build
