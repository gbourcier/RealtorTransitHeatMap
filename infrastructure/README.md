# Deployment

Single Docker image built by GitHub Actions, published to `ghcr.io/gbourcier/realtortransitheatmap`, pulled by a NAS via Watchtower.

## What runs where

- `db` — TimescaleDB, persisted to a NAS bind mount. **Not** auto-updated (major Postgres bumps need manual migration).
- `api` — Go binary serving both the REST API at `/api/*` and the Vue SPA at `/*`. Runs `migrate up` on every startup. Watchtower-enabled.

## NAS first-time setup

### 1. Prepare the host paths

The DB and the GTFS cache live on the NAS as bind mounts so they survive `compose down` and image updates.

```sh
DATA_DIR=/volume1/docker/realtor
mkdir -p "$DATA_DIR/db" "$DATA_DIR/gtfs-cache"
```

TimescaleDB runs as uid `999` (postgres) inside the container. The bind directory must be writable by that uid:

```sh
chown -R 999:999 "$DATA_DIR/db"
```

### 2. Copy the compose file and env

```sh
cd "$DATA_DIR"
curl -O https://raw.githubusercontent.com/gbourcier/RealtorTransitHeatMap/main/infrastructure/docker-compose.prod.yml
curl -o .env https://raw.githubusercontent.com/gbourcier/RealtorTransitHeatMap/main/infrastructure/.env.example
```

Edit `.env`. At minimum set:

- `DATA_DIR=/volume1/docker/realtor`
- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`
- `POLYGON_WKT` (search area, required)
- `GHCR_OWNER=gbourcier`
- `IMAGE_TAG=main` (Watchtower will track this tag)
- `API_PORT=3000` (or whatever the NAS exposes externally)

### 3. Bring it up

```sh
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

First boot will:

1. Start TimescaleDB, wait for healthcheck.
2. Run all migrations in `/app/migrations` against the DB.
3. Start the API on port 3000 inside the container.

Check `http://<nas-ip>:${API_PORT}/healthz` returns `ok` and `http://<nas-ip>:${API_PORT}/` serves the SPA.

### 4. Populate the transit graph (one-shot)

The GTFS precompute is a separate binary baked into the same image. Run it on demand:

```sh
docker compose -f docker-compose.prod.yml --env-file .env run --rm api /app/gtfs-precompute
```

It downloads agency feeds into `${DATA_DIR}/gtfs-cache` (cached across runs) and writes the `transit_stops` table. Rerun whenever transit schedules change.

## Watchtower

Run **one** Watchtower container for the whole NAS — it watches every container with the `com.centurylinklabs.watchtower.enable=true` label (the `api` service has it; `db` does not).

```yaml
services:
  watchtower:
    image: containrrr/watchtower
    container_name: watchtower
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      WATCHTOWER_LABEL_ENABLE: "true"
      WATCHTOWER_CLEANUP: "true"
      WATCHTOWER_POLL_INTERVAL: "300"
```

The image is published as a **public** GHCR package, so neither the NAS nor Watchtower needs `docker login` to pull it. If you ever flip the package back to private under GitHub Packages settings, you'll need to `docker login ghcr.io` on the NAS and mount `~/.docker/config.json` into the Watchtower container as `/config.json:ro`.

## Updating the DB image

Watchtower deliberately skips the DB. To upgrade Postgres/Timescale:

1. `docker compose ... exec db pg_dump ...` first.
2. Update the image tag in `docker-compose.prod.yml`.
3. `docker compose ... up -d db`.

Across **major** Postgres versions (e.g. pg16 → pg17), the data dir is incompatible and needs `pg_upgrade` or a dump/restore.

## Image tags

The CI workflow publishes:

- `:main` — every push to main (this is what Watchtower follows)
- `:sha-<short>` — every push, for rollback
- `:vX.Y.Z`, `:X.Y` — on git tags

To pin a known-good build instead of tracking `main`, set `IMAGE_TAG=sha-abc1234` and Watchtower will stop updating until you bump it again.
