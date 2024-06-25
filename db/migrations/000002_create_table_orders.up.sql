create table if not exists orders
(
    id         bigserial
        primary key,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    number     varchar,
    user_id    bigint,
    status     varchar not null default 'NEW',
    accrual    decimal(32, 2)
);

create index if not exists idx_orders_deleted_at
    on orders (deleted_at);

