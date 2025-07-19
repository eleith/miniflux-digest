FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main .

FROM alpine:3.22.1

WORKDIR /app

COPY --from=builder /app/main /app/main
COPY templates ./templates

ENTRYPOINT ["/app/main"]
