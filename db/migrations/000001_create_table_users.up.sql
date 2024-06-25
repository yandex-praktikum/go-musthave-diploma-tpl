create table if not exists users
(
    id          bigserial
        primary key,
    created_at  timestamp with time zone,
    updated_at  timestamp with time zone,
    deleted_at  timestamp with time zone,
    last_name   varchar,
    first_name  varchar,
    middle_name varchar,
    login       varchar not null
        constraint uni_users_login
            unique,
    password    varchar not null,
    email       varchar
);

create index if not exists idx_users_deleted_at
    on users (deleted_at);

