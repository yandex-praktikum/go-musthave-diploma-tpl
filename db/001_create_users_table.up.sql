CREATE TABLE IF NOT EXISTS users
(
    id       SERIAL PRIMARY KEY,
    login    VARCHAR,
    password VARCHAR,
    UNIQUE (login)
);
