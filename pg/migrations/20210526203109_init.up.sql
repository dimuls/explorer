create extension if not exists pg_trgm;
alter database explorer set pg_trgm.similarity_threshold = 0.3;

create table peer (
    id bigserial primary key,
    url text not null unique
);

create table channel (
     id bigserial primary key,
     name text not null unique
);

create table peer_channel (
    peer_id bigint not null references peer(id),
    channel_id bigint not null references channel(id),
    unique (peer_id, channel_id)
);

create table channel_config (
    id bigserial primary key,
    channel_id bigint not null references channel(id),
    raw bytea not null,
    parsed jsonb not null,
    created_at timestamp with time zone not null
);

create table chaincode (
    id bigserial primary key,
    name text not null,
    version text not null,
    unique (name, version)
);

create table channel_chaincode (
    channel_id bigint not null references channel(id),
    chaincode_id bigint not null references chaincode(id),
    unique (channel_id, chaincode_id)
);

create table block (
    id bigserial primary key,
    channel_id bigint not null references channel(id),
    number bigint not null,
    unique (channel_id, number)
);

create table transaction (
    id char(65) primary key,
    channel_id bigint not null references channel(id),
    block_id bigint not null references block (id),
    created_at timestamp with time zone not null
);

create table state (
    key text primary key,
    channel_id bigint not null references channel(id),
    transaction_id char(65) not null references transaction(id),
    type text not null,
    raw_value bytea not null,
    value jsonb,
    created_at timestamp with time zone not null
);

create table old_state (
    id bigserial primary key,
    channel_id bigint not null references channel(id),
    transaction_id char(65) not null references transaction(id),
    key text not null,
    type text not null,
    raw_value bytea not null,
    value jsonb,
    created_at timestamp with time zone not null
);