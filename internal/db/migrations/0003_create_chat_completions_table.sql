-- +goose up
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
