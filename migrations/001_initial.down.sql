-- ============================================================================
-- HP MMO BACKEND - ROLLBACK INITIAL MIGRATION
-- ============================================================================

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_equipment_updated_at ON equipment_loadouts;
DROP TRIGGER IF EXISTS trigger_inventory_updated_at ON inventory_items;
DROP TRIGGER IF EXISTS trigger_characters_updated_at ON characters;
DROP TRIGGER IF EXISTS trigger_accounts_updated_at ON accounts;
DROP TRIGGER IF EXISTS trigger_stats_version ON character_stats_snapshots;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at();
DROP FUNCTION IF EXISTS increment_stats_version();

-- Drop tables in reverse order
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS sanctions CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS zones CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS transactions_ledger CASCADE;
DROP TABLE IF EXISTS equipment_loadouts CASCADE;
DROP TABLE IF EXISTS inventory_items CASCADE;
DROP TABLE IF EXISTS item_definitions CASCADE;
DROP TABLE IF EXISTS character_currencies CASCADE;
DROP TABLE IF EXISTS currencies CASCADE;
DROP TABLE IF EXISTS character_flags CASCADE;
DROP TABLE IF EXISTS character_stats_snapshots CASCADE;
DROP TABLE IF EXISTS characters CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;

-- Drop types
DROP TYPE IF EXISTS audit_action CASCADE;
DROP TYPE IF EXISTS sanction_type CASCADE;
DROP TYPE IF EXISTS transaction_status CASCADE;
DROP TYPE IF EXISTS transaction_type CASCADE;
DROP TYPE IF EXISTS equipment_slot CASCADE;
DROP TYPE IF EXISTS item_type CASCADE;
DROP TYPE IF EXISTS character_status CASCADE;
DROP TYPE IF EXISTS role_type CASCADE;
DROP TYPE IF EXISTS house_type CASCADE;

-- Drop extensions
DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS pgcrypto;
DROP EXTENSION IF EXISTS "uuid-ossp";

DO $$
BEGIN
    RAISE NOTICE 'HP MMO Backend: Initial migration rolled back successfully!';
END $$;
