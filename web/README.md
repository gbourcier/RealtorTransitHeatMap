# Web (Vue + Vuetify)

Dashboard frontend for RealtorTransitHeatMap.

## Dev

```bash
cd web
npm install
npm run dev
```

Vite dev server runs on `:5173` and proxies `/api/*` to the Go API at
`127.0.0.1:3000`. Start the Go API separately (e.g. `go run ./cmd/api`).

## Build

```bash
npm run build
```
