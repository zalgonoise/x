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
    label_id INTEGER REFERENCES labels (id) NOT NULL,
    content 	TEXT 				NOT NULL,
	count 		INTEGER 			NOT NULL,
	ratio 		FLOAT 				NOT NULL
);
`
)
