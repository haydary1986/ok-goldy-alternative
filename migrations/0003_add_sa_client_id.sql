-- +goose Up
-- +goose StatementBegin

-- The Workspace Admin "Domain-Wide Delegation" page asks for the Service
-- Account's *Client ID* (the 21-digit Unique ID, distinct from the SA's
-- email). Having Goldy expose this value in the UI removes one of the
-- top sources of misconfiguration.
ALTER TABLE workspace_credentials
    ADD COLUMN IF NOT EXISTS sa_client_id TEXT;

-- Backfill the column from the JSON we already have on disk, so existing
-- deployments don't need a re-upload.
UPDATE workspace_credentials
SET sa_client_id = (convert_from(sa_json, 'UTF8')::jsonb) ->> 'client_id'
WHERE sa_client_id IS NULL
  AND sa_json IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE workspace_credentials DROP COLUMN IF EXISTS sa_client_id;
-- +goose StatementEnd
