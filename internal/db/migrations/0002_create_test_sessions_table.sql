-- +goose up
DROP TABLE IF EXISTS user_answers;

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

-- +goose down
DROP TABLE IF EXISTS user_answers;
DROP TABLE IF EXISTS test_sessions;

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
