BEGIN;

CREATE TABLE IF NOT EXISTS stats (
    peer_alias  bigint,
    user_id     varchar(255),
    version     varchar(255),
    created_at  timestamp DEFAULT now(),
    stats       json NOT NULL
);

CREATE INDEX idx_created_at ON stats(created_at);

COMMIT;
