# Ok Goldy Alternative

> A self-hosted, Go-native **alternative to [Ok Goldy](https://www.goldyarora.com/g-suite-addons/ok-goldy/)**, the popular but discontinued Google Sheets add-on for bulk Google Workspace administration.

If you used **Ok Goldy** to manage your Google Workspace from a spreadsheet and you're looking for a maintained replacement, this project is for you.

- **Built for scale:** designed for organizations with up to ~30,000 users.
- **Self-hosted:** runs in Docker; first-class **Coolify** support.
- **CSV-first:** drop-in workflow for existing Ok Goldy users.

## What it replaces

| Ok Goldy feature | Ok Goldy Alternative |
|---|---|
| Bulk **create / update / suspend / delete / export** users | ✅ |
| Bulk **create / delete / export** groups | ✅ |
| Bulk **add / remove / export** group members | ✅ |
| Bulk **add / remove / export** user aliases | ✅ |
| Spreadsheet-style row-and-column editing | ✅ (CSV import/export + Web UI) |
| Runs inside Google Sheets | ❌ — replaced by a self-hosted web app |

## Why a fresh project?

- The original Ok Goldy is **no longer maintained** by its author.
- Sheets-based add-ons are limited by Apps Script execution time (~6 min) and quotas — painful at scale.
- A standalone backend can manage **rate limiting, retries, and audit logging** properly.

## Architecture

```
┌──────────────┐   REST/JSON   ┌────────────────────┐
│ Web UI       │ ────────────► │  Go API Server     │
│ React + TS   │               │   chi + middleware  │
└──────────────┘               └─────┬───────┬──────┘
                                     │       │
                                ┌────▼───┐ ┌─▼────────┐
                                │Postgres│ │  Redis   │
                                │ jobs   │ │  asynq   │
                                │ audit  │ │  queue   │
                                └────────┘ └────┬─────┘
                                                │
                                          ┌─────▼──────┐
                                          │  Workers   │
                                          │ (rate-     │
                                          │  limited)  │
                                          └─────┬──────┘
                                                │
                                  Google Workspace Admin SDK
                                  (Service Account + DWD)
```

## Quick start

### 1. Prerequisites

- Google Workspace **Super Admin** account
- Google Cloud project with **Admin SDK API** enabled
- A **Service Account** key (JSON) granted **Domain-Wide Delegation** for these scopes:
  - `https://www.googleapis.com/auth/admin.directory.user`
  - `https://www.googleapis.com/auth/admin.directory.group`
  - `https://www.googleapis.com/auth/admin.directory.group.member`
  - `https://www.googleapis.com/auth/admin.directory.user.alias`

### 2. Local development

```bash
# 1. Configure environment
cp .env.example .env
mkdir -p secrets
cp /path/to/your-service-account.json secrets/service-account.json

# 2. Resolve Go modules
make deps

# 3. Bring up the stack (Postgres + Redis + migrate + server + worker)
make docker-up

# 4. Verify
curl http://localhost:8080/healthz
```

### 2b. Run the Web UI in dev

```bash
make web-install   # one-time
make web-dev       # http://localhost:5173 (proxies /api to the Go server on :8080)
```

### 3. Project layout

```
ok-goldy-alternative/
├── cmd/
│   ├── server/      # HTTP API entrypoint
│   ├── worker/      # asynq worker entrypoint
│   └── migrate/     # DB migration runner
├── internal/
│   ├── config/      # Env-based config loader
│   ├── log/         # slog logger setup
│   ├── db/          # pgx connection pool
│   ├── workspace/   # Google Admin SDK client (rate-limited)
│   ├── users/       # Users domain
│   ├── groups/      # Groups domain
│   ├── members/     # Group members domain
│   ├── aliases/     # User aliases domain
│   ├── jobs/        # asynq tasks
│   ├── audit/       # Audit log
│   ├── api/         # HTTP router, middleware
│   └── csv/         # CSV import / export
├── migrations/      # SQL migrations
├── web/             # React + TypeScript frontend
└── deploy/          # Coolify / k8s manifests
```

## Deploying to Coolify

1. In Coolify, create a new **Docker Compose** application and point it at this repository.
2. Copy every variable from `.env.example` into Coolify's environment variable editor and fill in real values.
3. Mount your `service-account.json` into the `secrets/` volume on the host.
4. Deploy. Coolify will build the image, run the `migrate` step, then start `server` and `worker`.

## Operations

### Scale considerations (30k users)

| Bottleneck | Mitigation |
|---|---|
| Admin SDK quota (default 2400 req / 100s) | Token-bucket rate limiter — `GOLDY_RATE_LIMIT_RPS` |
| Long-running bulk ops (~20+ min for 30k) | `asynq` queue with retries and resumability |
| Listing 30k users (~60 paginated calls) | Paginated fetch + local Postgres cache |
| Mid-job failure | Idempotent task design + per-row outcome audit |

### Observability

- Structured JSON logs via `log/slog`
- Per-request `X-Request-ID` middleware
- Audit log table records every mutation (actor, action, target, before/after)

## Roadmap

- [x] Project scaffold + Docker + Coolify-ready compose
- [x] Users domain (live CRUD; bulk + CSV pending)
- [x] Groups domain
- [x] Members domain
- [x] Aliases domain
- [x] Audit log (writes — read API pending)
- [x] React + TS Web UI scaffold (Vite + Tailwind + TanStack Query)
- [ ] Bulk operations + CSV import / export
- [ ] Audit log read API + UI
- [ ] Async job handlers (currently stubs)
- [ ] OIDC login (Google) for app admins
- [ ] OU / sub-org filtering
- [ ] go:embed the SPA build into the Go binary (single container)

## Acknowledgements

This project owes its name and inspiration to **[Goldy Arora](https://www.goldyarora.com/)** and the original **Ok Goldy** add-on, which served Workspace admins for many years.

## License

TBD
