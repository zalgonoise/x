CREATE TABLE items
(
    id           uuid PRIMARY KEY NOT NULL,
    image_source text,
    name         text             NOT NULL
);

CREATE TABLE ratings
(
    id      uuid PRIMARY KEY                             NOT NULL,
    item_id uuid REFERENCES items (id) ON DELETE CASCADE NOT NULL,
    user_id text                                         NOT NULL,
    points  int                                          NOT NULL
);

CREATE UNIQUE INDEX ratings_item_id_user_id ON ratings(item_id, user_id)
