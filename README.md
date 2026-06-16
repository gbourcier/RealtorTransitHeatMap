# Realtor Transit Heat Map

Realtor Transit Heat Map is a personal real estate dashboard for comparing
property listings by public-transit commute time. It imports listings from
Realtor.ca, precomputes commute estimates from GTFS feeds, and presents the
results in an authenticated map/list UI with filters, favorites, saved views,
and admin tools for scraper and transit-refresh jobs.

This project is built around Greater Montreal defaults: the bundled GTFS
refresh covers STM, STL, RTL, REM, and Exo feeds, and the default commute
target is near McGill. The target location, snapshot time, walking assumptions,
and search areas are configurable.

> This is an unofficial project and is not affiliated with Realtor.ca or any
> transit agency. If you run it yourself, make sure your use of external data
> sources complies with their terms and licenses.

## Features

- Authenticated Vue dashboard for browsing listings on a Leaflet map or in a
  resizable list panel.
- Listing filters for price, commute time, recency, bedrooms, bathrooms,
  interior area, building type, favorites, and expired listings.
- Saved filters and per-user favorite listings.
- Admin UI for recurring Realtor.ca scrape schedules, scrape run history, GTFS
  refresh schedules, transit refresh history, and user management.
- GTFS precompute pipeline that estimates travel time from transit stops to a
  configurable reference destination, then combines that with walking time from
  each listing to nearby stops.
- Go REST API serving both `/api/*` endpoints and the built Vue single-page app.
- Docker image and production compose file for self-hosting.

## Stack

- Backend: Go 1.25, chi, GORM, PostgreSQL/TimescaleDB, robfig/cron.
- Frontend: Vue 3, Vite, Vuetify, Pinia, Leaflet, TypeScript.
- Data: Realtor.ca listing responses plus public GTFS ZIP feeds.
- Tooling: Docker Compose, golang-migrate, npm, GitHub Actions, GHCR.

## Repository Layout

```text
cmd/api/               Go HTTP server
cmd/gtfs-precompute/   One-shot GTFS commute precompute command
internal/              Backend packages: auth, API, listings, scraper, transit
web/                   Vue/Vite frontend
migrations/            SQL migrations managed by golang-migrate
seeds/                 Optional local seed data
scripts/               Local setup, migration, and dev helpers
infrastructure/        Dockerfile, production compose, deployment notes
openapi.yaml           API reference draft
```

## Prerequisites

- Go 1.25+
- Node.js 20+ and npm
- Docker and Docker Compose
- `psql` client tools
- `golang-migrate` CLI with PostgreSQL support

The helper script `./scripts/setup-workspace.sh` installs `golang-migrate` if it
is missing, downloads Go modules, installs and builds the frontend, sets up the
database, builds the Go packages, and precomputes transit stops.

## Local Development

1. Create a local environment file:

   ```sh
   cp .env.example .env
   ```

2. Fill in the required values in `.env`. For local HTTP development, these are
   the usual minimums:

   ```dotenv
   POSTGRES_DB=realtor_heatmap
   POSTGRES_USER=realtor
   POSTGRES_PASSWORD=change-me
   POSTGRES_HOST=localhost
   POSTGRES_PORT=5432
   POSTGRES_ENABLE_SSL=false

   ADMIN_USERNAME=admin
   ADMIN_PASSWORD=change-me-too
   COOKIE_SECURE=false
   ```

   Optional but useful for local development:

   ```dotenv
   MOCK_REALTOR_API=true
   POSTGRES_SEED_DB=true
   ```

3. Start PostgreSQL/TimescaleDB:

   ```sh
   docker compose up -d --wait
   ```

4. Initialize the database:

   ```sh
   ./scripts/setup-db.sh
   ```

5. Install and build the frontend once:

   ```sh
   npm --prefix web install
   npm --prefix web run build
   ```

   The Go server embeds `web/dist`, so a fresh clone needs one frontend build
   before `make run` or `go test ./...`.

6. Start the backend:

   ```sh
   make run
   ```

7. In another terminal, start the frontend dev server:

   ```sh
   npm --prefix web run dev
   ```

The Vue dev server runs at `http://127.0.0.1:5173` and proxies `/api/*` to the
Go API at `http://127.0.0.1:3000`.

You can also run the combined helper:

```sh
./scripts/dev.sh
```

That script starts Docker Compose, the Go backend through `air`, and the Vite
frontend. Install `air` first if you want hot-reload for the backend.

## Transit Data

Transit commute data lives in the `transit_stops` table. To populate it locally:

```sh
go run ./cmd/gtfs-precompute
```

The command downloads GTFS ZIP files into `data/gtfs` by default, computes a
snapshot graph, and writes the resulting stop commute data to the database. On
later runs, use:

```sh
go run ./cmd/gtfs-precompute -skip-download
```

Useful flags:

- `-cache <dir>`: choose the GTFS cache directory.
- `-skip-download`: reuse cached feed ZIPs when present.
- `-dry-run`: compute without writing to the database.
- `-if-empty`: skip the run if `transit_stops` already has rows.

Transit defaults are configured with:

- `TRANSIT_SNAPSHOT` such as `tue-09`
- `TRANSIT_REF_LAT`, `TRANSIT_REF_LON`, `TRANSIT_REF_KEY`
- `TRANSIT_NEAREST_STOPS`
- `TRANSIT_WALK_SPEED_MPS`
- `TRANSIT_WALK_DETOUR`

## Common Commands

```sh
make test                    # go test ./...
make fmt                     # go fmt ./...
make vet                     # go vet ./...
make build                   # build bin/api
./scripts/migrate-up.sh      # apply pending migrations
./scripts/migrate-down.sh    # roll back one migration
npm --prefix web run build   # type-check and build the frontend
```

## Configuration

Required environment variables:

- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `ADMIN_USERNAME`
- `ADMIN_PASSWORD`

Frequently used optional variables:

- `HTTP_ADDR`: backend bind address, default `127.0.0.1:3000`.
- `POSTGRES_HOST`: default `localhost`.
- `POSTGRES_PORT`: default `5432`.
- `POSTGRES_ENABLE_SSL`: set to `false` for local Docker Postgres.
- `COOKIE_SECURE`: set to `false` for local HTTP; keep enabled behind HTTPS.
- `TRUST_PROXY`: enable only behind a trusted reverse proxy that overwrites
  `X-Forwarded-For`.
- `SESSION_TTL`: Go duration string, default `720h`.
- `REALTOR_BASE_URL`: default `https://api2.realtor.ca`.
- `MOCK_REALTOR_API`: set to `true` to use the checked-in mock response instead
  of calling Realtor.ca.
- `LAT_MAX`, `LAT_MIN`, `LON_MAX`, `LON_MIN`: optional bounding box used with
  per-schedule polygons.

See [.env.example](.env.example) and
[infrastructure/.env.example](infrastructure/.env.example) for the full list.

## Deployment

The production container is built by
[infrastructure/Dockerfile](infrastructure/Dockerfile). It packages:

- the Go API server,
- the built Vue app,
- the GTFS precompute binary,
- SQL migrations, and
- the `migrate` CLI.

On startup, [infrastructure/entrypoint.sh](infrastructure/entrypoint.sh) runs
database migrations, ensures transit data exists, and then starts the API.

The production compose file and NAS-oriented deployment notes live in
[infrastructure/](infrastructure/README.md). The GitHub Actions workflow builds
and publishes images to GHCR on pushes to `main`, version tags, and manual
workflow runs.

## Public Repo Notes

- Do not commit `.env`, database dumps, GTFS cache files, or private screenshots.
- Keep credentials, deployment hostnames, access tokens, and personal search
  criteria outside tracked files.
- The default `.gitignore` already excludes `.env`, build artifacts, and local
  `data/` caches.
- Treat upstream listing and GTFS data as third-party data. Respect rate limits,
  terms of service, and data licenses.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).
