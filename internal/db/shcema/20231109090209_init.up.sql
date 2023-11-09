CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    login varchar(50),
    password_hash varchar(255),
    salt varchar(255) not null,
    UNIQUE (login)
);

CREATE TABLE IF NOT EXISTS orders (
    id serial PRIMARY KEY, number VARCHAR(100) not null unique,
    status varchar(50),
    user_id int references users (id) on delete cascade not null,
    upload_date timestamp   DEFAULT now(),
    update_date timestamp  without time zone
) ;

CREATE TABLE IF NOT EXISTS balance (
    id serial PRIMARY KEY,
    number VARCHAR(100) not null,
    sum double precision not null DEFAULT 0,
    user_id int references users (id) on delete cascade not null,
    processed timestamp   DEFAULT now()
);
