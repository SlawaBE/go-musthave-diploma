CREATE SEQUENCE IF NOT EXISTS withdrawns_id_seq;

CREATE TABLE IF NOT EXISTS withdrawns (
    id BIGINT PRIMARY KEY DEFAULT nextval('withdrawns_id_seq'),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number VARCHAR(255) NOT NULL UNIQUE,
    total REAL NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_withdrawns_user_id ON withdrawns(user_id);

ALTER SEQUENCE withdrawns_id_seq OWNED BY withdrawns.id;
