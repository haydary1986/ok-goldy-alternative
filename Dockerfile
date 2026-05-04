# syntax=docker/dockerfile:1.7

# ---------- Frontend Builder ----------
FROM node:22-alpine AS frontend

WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm install --no-audit --no-fund

COPY web/ ./
RUN npm run build

# ---------- Go Builder ----------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .

# Replace the placeholder dist with the real built SPA.
RUN rm -rf web/dist
COPY --from=frontend /web/dist web/dist

# Re-tidy in case go.sum is missing (first build).
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server  ./cmd/server  \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker  ./cmd/worker  \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# ---------- Runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /out/server   /app/server
COPY --from=builder /out/worker   /app/worker
COPY --from=builder /out/migrate  /app/migrate
COPY --from=builder /src/migrations /app/migrations

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/app/server"]
