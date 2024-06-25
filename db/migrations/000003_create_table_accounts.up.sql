create table if not exists accounts
(
    id         bigserial
        primary key,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    type       varchar,
    sum        decimal(32, 2) not null default 0,
    user_id    bigint
);

create index if not exists idx_accounts_deleted_at
    on accounts (deleted_at);

insert into accounts (created_at, updated_at, type) values (now(), now(), 'system_withdraw');
