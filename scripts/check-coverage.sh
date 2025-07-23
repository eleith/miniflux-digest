#!/bin/sh

set -eu

if [ -z "$1" ]; then
    echo "Usage: $0 <coverage_threshold>"
    exit 1
fi

COVERAGE_THRESHOLD=$1

go test -mod=vendor -coverprofile=coverage.out ./...

COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

if ! echo "$COVERAGE" | grep -qE '^[0-9]+(\.[0-9]+)?$'; then
    echo "Error: Could not determine test coverage."
    exit 1
fi

echo "Total test coverage: $COVERAGE%"

if awk -v cov="$COVERAGE" -v min="$COVERAGE_THRESHOLD" 'BEGIN {exit (cov < min)}'; then
    echo "Success: Test coverage is at or above the ${COVERAGE_THRESHOLD}% threshold."
else
    echo "Error: Test coverage is below the ${COVERAGE_THRESHOLD}% threshold."
    exit 1
fi