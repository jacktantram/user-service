CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users
(
    id  UUID UNIQUE DEFAULT uuid_generate_v4(),
    first_name VARCHAR   NOT NULL,
    last_name VARCHAR NOT NULL,
    nickname VARCHAR NOT NULL,
    password VARCHAR NOT NULL,
    email  VARCHAR(100) NOT NULL,
    country VARCHAR(3) NOT NULL,
    created_at  timestamptz default now(),
    updated_at  timestamptz,

    UNIQUE(email)
);



