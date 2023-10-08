create table if not exists users (
    id serial primary key,
    name text not null,
    email text not null unique,
    password text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

create table if not exists permissions (
    id serial primary key,
    name text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

create table if not exists organizations (
    id serial primary key,
    name text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

create table if not exists api_keys (
    id serial primary key,
    name text not null,
    key text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

create table if not exists settings (
    id serial primary key,
    name text not null,
    value text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

create table if not exists tasks (
    id serial primary key,
    name text not null,
    description text not null,
    status text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);