CREATE TABLE sites (
   id UUID PRIMARY KEY,
   enabled BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE api_tokens (
    id UUID PRIMARY KEY,
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    token VARCHAR(68) NOT NULL UNIQUE,
    revoked_at TIMESTAMPTZ NULL
);

CREATE INDEX api_tokens_active_lookup_idx
    ON api_tokens (token)
    WHERE revoked_at IS NULL;