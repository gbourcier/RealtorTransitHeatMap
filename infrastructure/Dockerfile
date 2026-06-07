FROM node:20-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.25-alpine AS api
RUN apk add --no-cache git
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1
COPY . .
COPY --from=web /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/gtfs-precompute ./cmd/gtfs-precompute \
 && cp "$(go env GOPATH)/bin/migrate" /out/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget \
 && addgroup -S app && adduser -S -G app app \
 && mkdir -p /app/data/gtfs \
 && chown -R app:app /app
WORKDIR /app
COPY --from=api /out/api /app/api
COPY --from=api /out/gtfs-precompute /app/gtfs-precompute
COPY --from=api /out/migrate /app/migrate
COPY migrations /app/migrations
COPY infrastructure/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh /app/api /app/gtfs-precompute /app/migrate
USER app
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -q -O- http://127.0.0.1:3000/healthz >/dev/null 2>&1 || exit 1
ENTRYPOINT ["/app/entrypoint.sh"]
