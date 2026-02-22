FROM golang:1.26.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV GOEXPERIMENT=goroutineleakprofile
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
    -o ./server ./cmd/server

# Final stage
FROM alpine:3.21
WORKDIR /app
RUN apk update && apk upgrade && \
    apk add --no-cache ca-certificates tzdata curl && \
    rm -rf /var/cache/apk/*
COPY --from=builder /app/server ./server
COPY --from=builder /app/config ./config
ENTRYPOINT ["./server"]
