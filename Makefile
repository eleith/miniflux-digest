build:
	go build -mod=vendor -o miniflux-digest ./cmd/miniflux-digest

preview-html:
ifdef minify
	go run -mod=vendor ./scripts/preview-html/main.go -minify=${minify}
else
	go run -mod=vendor ./scripts/preview-html/main.go
endif
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
