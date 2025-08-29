-- Remove UNIQUE constraint from ip_address in connections table
-- This allows multiple connections to use the same IP address (e.g., different VLANs, ports, etc.)

-- First, check if the constraint exists and drop it
DO $$ 
BEGIN
    -- Drop unique constraint on ip_address if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'connections_ip_address_key' 
        AND table_name = 'connections'
    ) THEN
        ALTER TABLE connections DROP CONSTRAINT connections_ip_address_key;
        RAISE NOTICE 'Dropped unique constraint on connections.ip_address';
    ELSE
        RAISE NOTICE 'Unique constraint on connections.ip_address does not exist';
    END IF;
END $$;