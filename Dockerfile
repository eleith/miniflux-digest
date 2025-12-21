FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY web/static ./web/static

RUN go mod vendor
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o /app/miniflux-digest ./cmd/miniflux-digest

FROM alpine:3.22.1

WORKDIR /app

COPY --from=builder /app/miniflux-digest /app/miniflux-digest
COPY --from=builder /app/web/static /app/web/static
ENTRYPOINT ["/app/miniflux-digest"]
