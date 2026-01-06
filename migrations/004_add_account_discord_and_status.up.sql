-- 004: Add discord_id and status to accounts table

-- Create account status enum
CREATE TYPE account_status AS ENUM ('active', 'whitelisted', 'banned', 'kicked_out');

-- Add new columns to accounts table
ALTER TABLE accounts
    ADD COLUMN discord_id VARCHAR(32),
    ADD COLUMN status account_status NOT NULL DEFAULT 'active';

-- Create index on discord_id for lookups
CREATE INDEX idx_accounts_discord_id ON accounts(discord_id) WHERE discord_id IS NOT NULL AND deleted_at IS NULL;

-- Create index on status for filtering
CREATE INDEX idx_accounts_status ON accounts(status) WHERE deleted_at IS NULL;

-- Add unique constraint on discord_id
ALTER TABLE accounts
    ADD CONSTRAINT accounts_discord_id_unique UNIQUE (discord_id);
