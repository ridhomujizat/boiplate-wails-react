-- Create app_settings table for storing application configuration
CREATE TABLE IF NOT EXISTS app_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key VARCHAR(255) NOT NULL UNIQUE,
    value TEXT,
    type VARCHAR(50) DEFAULT 'string',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster key lookups
CREATE INDEX IF NOT EXISTS idx_app_settings_key ON app_settings(key);

-- Insert default settings
INSERT OR IGNORE INTO app_settings (key, value, type) VALUES
('tenant', NULL, 'string'),
('baseurl', NULL, 'string'),
('mqtt', NULL, 'string');