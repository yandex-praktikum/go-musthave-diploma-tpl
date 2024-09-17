create table if not exists "user"
(
    id            bigserial
        constraint user_pk
            primary key,
    login         varchar(100) not null,
    password_hash varchar(255) not null
);

create unique index if not exists user_login_uindex
    on "user" (login);

create table if not exists orders
(
    id         bigserial
        constraint orders_pk
            primary key,
    number     varchar(32)                               not null,
    status     smallint                                  not null,
    user_id    bigserial
        constraint orders_user_id_fk
            references "user"
            on delete cascade,
    created_at timestamp(3) with time zone default now() not null
);

create unique index if not exists orders_number_uindex
    on orders (number);

create table if not exists balance
(
    id         bigserial
        constraint id
            primary key,
    order_id   bigserial
        constraint balance_orders_id_fk
            references orders
            on delete cascade,
    sum        bigint                                    not null,
    type       smallint                                  not null,
    created_at timestamp(3) with time zone default now() not null
);

create unique index if not exists balance_order_id_type_uindex
    on balance (order_id, type)
    where (type = 0);
