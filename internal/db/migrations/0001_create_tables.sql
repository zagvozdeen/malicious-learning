-- +goose up
CREATE TABLE IF NOT EXISTS telegram_updates
(
    id     BIGINT PRIMARY KEY,
    update JSON        NOT NULL,
    date   TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS users
(
    id         SERIAL PRIMARY KEY,
    tid        BIGINT       NULL UNIQUE,
    uuid       UUID         NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    last_name  VARCHAR(255) NULL,
    username   VARCHAR(255) NULL,
    email      VARCHAR(256) NULL,
    password   VARCHAR(256) NULL,
    created_at TIMESTAMPTZ  NOT NULL,
    updated_at TIMESTAMPTZ  NOT NULL
);

CREATE TABLE IF NOT EXISTS modules
(
    id         SERIAL PRIMARY KEY,
    uuid       UUID         NOT NULL UNIQUE,
    name       VARCHAR(256) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL,
    updated_at TIMESTAMPTZ  NOT NULL
);

CREATE TABLE IF NOT EXISTS cards
(
    id         SERIAL PRIMARY KEY,
    uid        INTEGER     NOT NULL,
    uuid       UUID        NOT NULL UNIQUE,
    question   TEXT        NOT NULL,
    answer     TEXT        NOT NULL,
    module_id  INTEGER     NOT NULL REFERENCES modules (id) ON DELETE RESTRICT,
    is_active  BOOLEAN     NOT NULL,
    hash       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TYPE user_answer_status AS ENUM ('null', 'remember', 'forgot');

CREATE TABLE IF NOT EXISTS user_answers
(
    id         SERIAL PRIMARY KEY,
    uuid       UUID               NOT NULL UNIQUE,
    group_uuid UUID               NOT NULL,
    card_id    INTEGER            NOT NULL REFERENCES cards (id) ON DELETE CASCADE,
    user_id    INTEGER            NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status     user_answer_status NOT NULL,
    created_at TIMESTAMPTZ        NOT NULL,
    updated_at TIMESTAMPTZ        NOT NULL
);

-- +goose down
DROP TABLE IF EXISTS user_answers;
DROP TYPE IF EXISTS user_answer_status;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS telegram_updates;
