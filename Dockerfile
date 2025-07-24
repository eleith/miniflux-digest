FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod vendor

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o /app/miniflux-digest ./cmd/miniflux-digest

FROM alpine:3.22.1

WORKDIR /app

COPY --from=builder /app/miniflux-digest /app/miniflux-digest
ENTRYPOINT ["/app/miniflux-digest"]
