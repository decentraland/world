BEGIN;

CREATE TABLE IF NOT EXISTS stats (
    peer_alias  bigint,
    version     varchar(255),
    created_at  timestamp DEFAULT now(),
    stats       json NOT NULL
);

COMMIT;
