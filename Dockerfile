FROM golang:alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o listmonk .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata shadow su-exec
WORKDIR /listmonk
COPY --from=builder /build/listmonk .
COPY config.toml.sample config.toml
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
EXPOSE 9000
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["./listmonk"]
