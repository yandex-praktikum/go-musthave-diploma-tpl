create table orders
(
    id         bigserial
        primary key,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    number     varchar,
    user_id    bigint,
    status     text,
    accrual    bigint
);

alter table orders
    owner to application;

create index idx_orders_deleted_at
    on orders (deleted_at);

