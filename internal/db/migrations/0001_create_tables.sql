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

CREATE TABLE IF NOT EXISTS courses
(
    id         SERIAL PRIMARY KEY,
    uuid       UUID         NOT NULL UNIQUE,
    slug       VARCHAR(256) NOT NULL UNIQUE,
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
    course_id  INTEGER     NOT NULL REFERENCES courses (id) ON DELETE RESTRICT,
    is_active  BOOLEAN     NOT NULL,
    hash       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS test_sessions
(
    id              SERIAL PRIMARY KEY,
    uuid            UUID        NOT NULL UNIQUE,
    user_id         INTEGER     NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    module_ids      INTEGER[]   NOT NULL,
    is_shuffled     BOOLEAN     NOT NULL,
    is_active       BOOLEAN     NOT NULL,
    recommendations TEXT        NULL,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL
);

CREATE TYPE user_answer_status AS ENUM ('null', 'remember', 'forgot');

CREATE TABLE IF NOT EXISTS user_answers
(
    id              SERIAL PRIMARY KEY,
    uuid            UUID               NOT NULL UNIQUE,
    card_id         INTEGER            NOT NULL REFERENCES cards (id) ON DELETE CASCADE,
    test_session_id INTEGER            NOT NULL REFERENCES test_sessions (id) ON DELETE CASCADE,
    status          user_answer_status NOT NULL,
    created_at      TIMESTAMPTZ        NOT NULL,
    updated_at      TIMESTAMPTZ        NOT NULL
);

CREATE TABLE IF NOT EXISTS chat_completions
(
    id                SERIAL PRIMARY KEY,
    uuid              UUID         NOT NULL UNIQUE,
    test_session_id   INTEGER      NOT NULL REFERENCES test_sessions (id) ON DELETE CASCADE,
    model             VARCHAR(255) NOT NULL,
    completion_tokens BIGINT       NOT NULL,
    prompt_tokens     BIGINT       NOT NULL,
    total_tokens      BIGINT       NOT NULL,
    date              BIGINT       NOT NULL,
    created_at        TIMESTAMPTZ  NOT NULL,
    updated_at        TIMESTAMPTZ  NOT NULL
);

-- +goose down
DROP TABLE IF EXISTS chat_completions;
DROP TABLE IF EXISTS user_answers;
DROP TYPE IF EXISTS user_answer_status;
DROP TABLE IF EXISTS test_sessions;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS telegram_updates;
