package database

const (
	createLabelsTable = `
CREATE TABLE labels 
(
    id      INTEGER PRIMARY KEY NOT NULL,
    label 	TEXT 				NOT NULL
);
`

	createLabelItemsTable = `
CREATE TABLE label_items
(
    id      INTEGER PRIMARY KEY NOT NULL,
    label_id INTEGER REFERENCES labels (id) NOT NULL,
    content 	TEXT 				NOT NULL,
	count 		INTEGER 			NOT NULL,
	ratio 		FLOAT 				NOT NULL
);
`

	createPlaylistsTable = `
CREATE TABLE playlists 
(
    id      INTEGER PRIMARY KEY NOT NULL,
		label_id INTEGER REFERENCES labels (id) NOT NULL
);`

	createPlaylistItemsTable = `
CREATE TABLE playlist_items
(
    id INTEGER PRIMARY KEY NOT NULL,
    playlist_id INTEGER REFERENCES playlists (id) NOT NULL,
    item_id INTEGER REFERENCES items (id) NOT NULL    
);`
)
