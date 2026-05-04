-- +goose Up
-- +goose StatementBegin

-- Singleton table holding the Google Workspace service-account credentials.
-- Uploaded by an admin through the Settings page; supersedes the
-- GOLDY_GOOGLE_* environment variables when present.
--
-- The id CHECK enforces a single row, so an UPSERT on id=1 is the only
-- legal write path.
CREATE TABLE IF NOT EXISTS workspace_credentials (
    id              SMALLINT     PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    sa_json         BYTEA        NOT NULL,
    delegated_admin TEXT         NOT NULL,
    customer_id     TEXT         NOT NULL DEFAULT 'my_customer',
    sa_email        TEXT,
    project_id      TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE workspace_credentials IS
    'Service-account JSON + delegated-admin config for Google Workspace API. Singleton (id=1).';
COMMENT ON COLUMN workspace_credentials.sa_json IS
    'Raw service-account JSON key bytes. Stored unencrypted; rely on filesystem and DB-level access controls.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workspace_credentials;
-- +goose StatementEnd
