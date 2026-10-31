CREATE TABLE links (
    code    TEXT PRIMARY KEY,
    url     TEXT NOT NULL,
    hits    INTEGER NOT NULL DEFAULT 0,
    created INTEGER NOT NULL -- Unix-Nanosekunden, UTC
) STRICT;
