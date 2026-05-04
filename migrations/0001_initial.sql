-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- High-level bulk job records driven by the API and consumed by the worker.
CREATE TABLE IF NOT EXISTS jobs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type          TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    total_rows    INT NOT NULL DEFAULT 0,
    success_rows  INT NOT NULL DEFAULT 0,
    failure_rows  INT NOT NULL DEFAULT 0,
    payload       JSONB,
    actor         TEXT NOT NULL DEFAULT '',
    started_at    TIMESTAMPTZ,
    finished_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_type   ON jobs(type);

-- Per-row outcome of a bulk job. Enables resumability and detailed reporting.
CREATE TABLE IF NOT EXISTS job_rows (
    id            BIGSERIAL PRIMARY KEY,
    job_id        UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    row_index     INT NOT NULL,
    target        TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    error_code    TEXT,
    error_message TEXT,
    payload       JSONB,
    processed_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_job_rows_job    ON job_rows(job_id);
CREATE INDEX IF NOT EXISTS idx_job_rows_status ON job_rows(status);

-- Audit log: one row per mutating action, ever.
CREATE TABLE IF NOT EXISTS audit_log (
    id             BIGSERIAL PRIMARY KEY,
    occurred_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor          TEXT NOT NULL,
    action         TEXT NOT NULL,
    resource_type  TEXT NOT NULL,
    resource_id    TEXT NOT NULL,
    job_id         UUID REFERENCES jobs(id) ON DELETE SET NULL,
    request_id     TEXT,
    before         JSONB,
    after          JSONB,
    ok             BOOLEAN NOT NULL DEFAULT TRUE,
    error_message  TEXT
);
CREATE INDEX IF NOT EXISTS idx_audit_occurred ON audit_log(occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_actor    ON audit_log(actor);
CREATE INDEX IF NOT EXISTS idx_audit_resource ON audit_log(resource_type, resource_id);

-- Local cache of Workspace users to avoid re-listing 30k records on every page.
CREATE TABLE IF NOT EXISTS users_cache (
    id            TEXT PRIMARY KEY,
    primary_email TEXT UNIQUE NOT NULL,
    given_name    TEXT,
    family_name   TEXT,
    org_unit_path TEXT,
    suspended     BOOLEAN NOT NULL DEFAULT FALSE,
    is_admin      BOOLEAN NOT NULL DEFAULT FALSE,
    raw           JSONB,
    synced_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_users_cache_ou        ON users_cache(org_unit_path);
CREATE INDEX IF NOT EXISTS idx_users_cache_suspended ON users_cache(suspended);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users_cache;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS job_rows;
DROP TABLE IF EXISTS jobs;
-- +goose StatementEnd
