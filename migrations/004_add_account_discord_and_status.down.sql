-- Rollback: Remove discord_id and status from accounts

-- Drop indexes
DROP INDEX IF EXISTS idx_accounts_discord_id;
DROP INDEX IF EXISTS idx_accounts_status;

-- Remove columns
ALTER TABLE accounts
    DROP COLUMN IF EXISTS discord_id,
    DROP COLUMN IF EXISTS status;

-- Drop enum type
DROP TYPE IF EXISTS account_status;
