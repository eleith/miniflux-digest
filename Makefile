# Build the application binary using vendored dependencies.
build:
	go build -mod=vendor -o miniflux-digest .

# Run all tests using vendored dependencies.
test:
	go test -mod=vendor ./...

# Run the linter. It will automatically use the .golangci.yml config.
lint:
	golangci-lint run

# Create the vendor directory.
vendor:
	go mod vendor

# ci is a target to run all the checks that the CI server runs.
ci: vendor lint test

# The default command, running 'make' will run all CI checks and build the binary.
all: ci build
