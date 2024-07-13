CREATE TABLE gophermart.users (
    id SERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    wallet DECIMAL NOT NULL DEFAULT 0,
    withdrawn DECIMAL NOT NULL DEFAULT 0
);