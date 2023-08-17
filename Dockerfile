FROM golang:1.20 AS builder

WORKDIR /app

COPY . .

RUN --mount=type=cache,mode=0755,target=/go/pkg/mod go mod tidy
RUN --mount=type=cache,mode=0755,target=/go/pkg/mod CGO_ENABLED=0 GOOS=linux go build -o /server-app ./cmd/server

FROM alpine:latest

COPY --from=builder /server-app /server-app

ENTRYPOINT [ "/server-app" ]
