CREATE TYPE file_status AS ENUM ('pending', 'ready', 'quarantined', 'deleted');

CREATE TABLE files (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    VARCHAR(255) NOT NULL,
    name        VARCHAR(1024) NOT NULL,
    size        BIGINT NOT NULL,
    mime_type   VARCHAR(255) NOT NULL,
    status      file_status NOT NULL DEFAULT 'pending',
    storage_key VARCHAR(1024),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_files_owner_id ON files(owner_id);
CREATE INDEX idx_files_status ON files(status);
