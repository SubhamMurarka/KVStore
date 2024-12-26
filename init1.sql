-- Create the replicator user with replication privileges
CREATE USER replicator WITH REPLICATION ENCRYPTED PASSWORD 'replicator_password';

-- Create the physical replication slot
SELECT pg_create_physical_replication_slot('replication_slot');

-- Create the kv_store table
CREATE TABLE IF NOT EXISTS kv_store (
    key VARCHAR(255) PRIMARY KEY, 
    value TEXT,          
    expire_at TIMESTAMP            
);

-- Grant appropriate privileges to the replicator user
GRANT CONNECT ON DATABASE postgres TO replicator;
GRANT USAGE ON SCHEMA public TO replicator;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO replicator;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO replicator;

-- Create indexes on the kv_store table for optimized queries
CREATE INDEX IF NOT EXISTS idx_key ON kv_store (key);
CREATE INDEX IF NOT EXISTS idx_expire_at ON kv_store (expire_at);
