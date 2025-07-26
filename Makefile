build:
	go build -mod=vendor -o miniflux-digest ./cmd/miniflux-digest

# Define default values for preview flags using ?=
# These values are used if not explicitly overridden on the 'make' command line.
MINIFY ?= true
GROUP_BY ?= day

preview-html:
	go run -mod=vendor ./scripts/preview-html/main.go -minify=$(MINIFY) -group-by=$(GROUP_BY)
	./scripts/open-preview.sh

preview-miniflux-email:
	go run -mod=vendor ./scripts/preview-miniflux-email/main.go ${id}

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