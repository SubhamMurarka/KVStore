CREATE TABLE IF NOT EXISTS kv_store (
    key VARCHAR(255) PRIMARY KEY, 
    value TEXT,          
    expire_at TIMESTAMP            
);

CREATE INDEX IF NOT EXISTS idx_key ON kv_store (key);

CREATE INDEX IF NOT EXISTS idx_expire_at ON kv_store (expire_at);

