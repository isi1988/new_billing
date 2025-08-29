-- Migration for traffic data table

CREATE TABLE IF NOT EXISTS traffic (
    id SERIAL PRIMARY KEY,
    connection_id INTEGER NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    bytes_in BIGINT NOT NULL DEFAULT 0,
    bytes_out BIGINT NOT NULL DEFAULT 0,
    packets_in BIGINT NOT NULL DEFAULT 0,
    packets_out BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_traffic_client_id ON traffic(client_id);
CREATE INDEX IF NOT EXISTS idx_traffic_connection_id ON traffic(connection_id);
CREATE INDEX IF NOT EXISTS idx_traffic_timestamp ON traffic(timestamp);
CREATE INDEX IF NOT EXISTS idx_traffic_client_timestamp ON traffic(client_id, timestamp);

-- Insert sample data for testing
INSERT INTO traffic (connection_id, client_id, timestamp, bytes_in, bytes_out, packets_in, packets_out) 
SELECT 
    c.id as connection_id,
    ct.client_id,
    NOW() - INTERVAL '1 hour' * (random() * 24)::integer,
    (random() * 1000000000)::bigint,
    (random() * 500000000)::bigint,
    (random() * 1000000)::bigint,
    (random() * 800000)::bigint
FROM connections c
JOIN contracts ct ON c.contract_id = ct.id
ORDER BY random()
LIMIT 50;