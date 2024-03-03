CREATE TABLE services
(
    id           INTEGER PRIMARY KEY NOT NULL,
    name         TEXT                NOT NULL UNIQUE,
    pub_key      BLOB                NOT NULL,
    cert         BLOB                NULL
);


CREATE TABLE challenges
(
    service_id    INTEGER REFERENCES services (id) NOT NULL,
    challenge     BLOB                             NULL,
    expiry        INTEGER                          NULL
);

CREATE TABLE tokens
(
    service_id    INTEGER REFERENCES services (id) NOT NULL,
    token         BLOB                             NULL,
    expiry        INTEGER                          NULL
);

CREATE UNIQUE INDEX idx_challenges_service_id ON challenges (service_id);
CREATE UNIQUE INDEX idx_tokens_service_id ON tokens (service_id);
