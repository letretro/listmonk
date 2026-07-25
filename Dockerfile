FROM node:20-alpine AS frontend
WORKDIR /build
COPY frontend/package.json frontend/yarn.lock ./
RUN yarn install --frozen-lockfile
COPY frontend/ .
RUN yarn build

FROM golang:alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /build/dist ./frontend/dist/
RUN CGO_ENABLED=0 go build -o listmonk ./cmd/

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata shadow su-exec
WORKDIR /listmonk
COPY --from=builder /build/listmonk .
COPY --from=builder /build/config.toml.sample config.toml
COPY --from=builder /build/queries ./queries/
COPY --from=builder /build/schema.sql .
COPY --from=builder /build/permissions.json .
COPY --from=builder /build/frontend/dist ./frontend/dist/
COPY --from=builder /build/static ./static/
COPY --from=builder /build/i18n ./i18n/
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
EXPOSE 9000
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["./listmonk"]
