CREATE TABLE services
(
    id           INTEGER PRIMARY KEY NOT NULL,
    name         TEXT                NOT NULL UNIQUE,
    pub_key      BLOB                NOT NULL,
    cert         BLOB                NULL
);