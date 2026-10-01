CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE SEQUENCE IF NOT EXISTS orders_id_seq;

CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY DEFAULT nextval('orders_id_seq'),
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    number VARCHAR(255) NOT NULL UNIQUE,
    status order_status NOT NULL DEFAULT 'NEW',
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

ALTER SEQUENCE orders_id_seq OWNED BY orders.id;
