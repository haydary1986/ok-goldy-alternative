# Ok Goldy Alternative — Web UI

React + TypeScript SPA built with Vite, Tailwind CSS, and TanStack Query.

## Local development

```bash
# from /web
npm install
npm run dev
```

The dev server runs on `http://localhost:5173` and proxies `/api/*`,
`/healthz`, and `/readyz` to the Go server on `http://localhost:8080`.
Make sure the Go server is running (e.g. `make run` from the project root).

## Production build

```bash
npm run build       # outputs to web/dist
npm run preview     # serves the production build locally
```

For deployment, serve `web/dist` from any static host (nginx, Caddy, Coolify
static service, S3 + CloudFront, …). The SPA expects API routes at the same
origin under `/api/v1` — point a reverse proxy at the Go server, or run them
behind the same Coolify proxy.

## Identifying the actor

Until OIDC login lands, every authenticated request uses an actor email read
from `localStorage` (`goldy_actor`). The Layout header has an email input you
can edit; the value is sent as the `X-Goldy-Actor` header so the Go server's
audit log records the correct user.
