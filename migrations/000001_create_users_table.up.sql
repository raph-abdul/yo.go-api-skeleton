-- Licensed under the Apache License, Version 2.0. See LICENSE file.
CREATE TABLE IF NOT EXISTS users
(
    id
    UUID
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
),
    name VARCHAR
(
    100
) NOT NULL,
    email VARCHAR
(
    255
) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL, -- Store hashed passwords only!
    role VARCHAR
(
    50
) NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW
(
),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW
(
)
    );

-- Optional: Add index for faster email lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);