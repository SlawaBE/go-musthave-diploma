ALTER TABLE users ADD COLUMN id BIGINT;

WITH numbered AS (
    SELECT 
        login,
        ROW_NUMBER() OVER (ORDER BY login) AS new_id
    FROM users
)
UPDATE users 
SET id = numbered.new_id
FROM numbered 
WHERE users.login = numbered.login;

ALTER TABLE users ALTER COLUMN id SET NOT NULL;

CREATE SEQUENCE IF NOT EXISTS users_id_seq;
SELECT setval('users_id_seq', (SELECT COALESCE(MAX(id), 1) FROM users));

ALTER TABLE users ALTER COLUMN id SET DEFAULT nextval('users_id_seq');

ALTER TABLE users DROP CONSTRAINT users_pkey;

ALTER TABLE users ADD PRIMARY KEY (id);

ALTER TABLE users ADD CONSTRAINT users_login_unique UNIQUE (login);
