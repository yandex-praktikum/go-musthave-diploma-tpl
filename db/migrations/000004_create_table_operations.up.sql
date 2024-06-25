create table if not exists operations
(
    id                   bigserial
        primary key,
    created_at           timestamp with time zone,
    updated_at           timestamp with time zone,
    deleted_at           timestamp with time zone,
    processed_at         timestamp with time zone,
    type                 varchar,
    order_number         varchar,
    sum                  decimal(32, 2),
    sender_account_id    bigint,
    recipient_account_id bigint
);

create index if not exists idx_operations_deleted_at
    on operations (deleted_at);

