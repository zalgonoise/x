CREATE TABLE IF NOT EXISTS traces (
    traceID BLOB NOT NULL,
    timestamp INTEGER NOT NULL,
    UNIQUE(traceID)
);
