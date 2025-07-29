build:
	go build -mod=vendor -o miniflux-digest ./cmd/miniflux-digest

preview-html:
	go run -mod=vendor ./scripts/preview/main.go

preview-email:
	go run -mod=vendor ./scripts/preview/main.go --email

preview-miniflux:
ifdef category
	go run -mod=vendor ./scripts/preview/main.go --miniflux=${category}
else
	@echo "use 'preview-miniflux category=' to preview a category from miniflux"
endif

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
