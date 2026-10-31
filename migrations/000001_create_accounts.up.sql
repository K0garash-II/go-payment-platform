CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    owner VARCHAR(50) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
);