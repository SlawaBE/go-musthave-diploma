CREATE TABLE IF NOT EXISTS users (
    login VARCHAR(255) PRIMARY KEY,
    password_hash VARCHAR(255)
);
