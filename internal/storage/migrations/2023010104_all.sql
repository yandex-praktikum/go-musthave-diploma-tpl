-- +goose Up
-- +goose StatementBegin

-- users table

create table if not exists users (
    id serial primary key,
    login text not null,
    password text not null
    -- constraint users_pkey primary key (id)
);

create unique index if not exists idx_users_login ON users USING btree (login);

-- orders table

create table if not exists orders (
    id numeric not null,
    amount numeric null,
    user_id text not null,
    cr_dt timestamptz not null,
    status text not null,
    constraint orders_pkey primary key (id)
);

-- withdrawals table

create table if not exists withdrawals (
    id serial primary key,
    order_id numeric not null,
    user_id text not null,
    total numeric not null,
    cr_dt timestamptz not null,
    status text not null
);

-- +goose StatementEnd

-- +goose Down
drop table if exists users;
drop table if exists orders;
drop table if exists withdrawals;