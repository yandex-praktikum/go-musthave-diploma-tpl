BEGIN;
------------
-- TABLES --
------------
CREATE TABLE IF NOT EXISTS users
(
    id            serial
        constraint users_pk
            primary key,
    registered_at timestamp with time zone not null,
    login         varchar   not null,
    password      varchar   not null,
    last_sign_in  timestamp with time zone not null
);
create unique index users_login_uindex
    on users (login);


CREATE TYPE STATS AS enum ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE IF NOT EXISTS orders
(
    number      varchar       not null,
    user_id     int           not null,
    uploaded_at timestamp with time zone not null,
    status      STATS         not null,
    accrual     numeric default 0 not null
);
create unique index orders_number_uindex
    on orders (number);


CREATE TABLE IF NOT EXISTS balances
(
    id           serial
    constraint balances_pk
    primary key,
    user_id      int           not null,
    processed_at timestamp with time zone not null,
    income       numeric default 0 not null,
    outcome      numeric default 0 not null,
    order_number varchar       not null
);

---------------
-- FUNCTIONS --
---------------

----------
-- DATA --
----------


COMMIT;