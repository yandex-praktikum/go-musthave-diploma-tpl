package store

const createURLTable = `
		CREATE TABLE IF NOT EXISTS urls (
			primary_id integer PRIMARY KEY,
			id varchar36 UNIQUE,
			login varchar255 UNIQUE,
		    password_hash varchar60,
		    created_at	datetime NOT NULL
		)
	`
