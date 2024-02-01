CREATE TABLE entities
(
    id           INTEGER PRIMARY KEY NOT NULL,
    name         TEXT                NOT NULL UNIQUE,
    pub_key      BLOB                NOT NULL,
    cert         BLOB                NULL
);

CREATE TABLE challenges
(
    entity_id    INTEGER REFERENCES entities (id) NOT NULL,
    expiry       INTEGER             NOT NULL,
    challenge    BLOB                NOT NULL
);

CREATE TABLE tokens
(
    entity_id    INTEGER REFERENCES entities (id) NOT NULL,
    token        TEXT                             NULL,
    expiry       INTEGER                          NULL
);

CREATE UNIQUE INDEX idx_challenges_entity_id ON challenges (entity_id);
CREATE UNIQUE INDEX idx_tokens_entity_id ON tokens (entity_id);
