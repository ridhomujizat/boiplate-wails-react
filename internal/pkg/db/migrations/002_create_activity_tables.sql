-- Create activity_events table for storing user activity data
CREATE TABLE IF NOT EXISTS activity_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_name VARCHAR(255) NOT NULL,
    bundle_id VARCHAR(255),
    window_title VARCHAR(512),
    tab_title VARCHAR(512),
    url VARCHAR(1024),
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    duration INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    os VARCHAR(20),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for faster queries
CREATE INDEX IF NOT EXISTS idx_activity_events_app_name ON activity_events(app_name);
CREATE INDEX IF NOT EXISTS idx_activity_events_start_time ON activity_events(start_time);
CREATE INDEX IF NOT EXISTS idx_activity_events_end_time ON activity_events(end_time);
CREATE INDEX IF NOT EXISTS idx_activity_events_status ON activity_events(status);

-- Create aggregated_stats table for daily aggregated statistics
CREATE TABLE IF NOT EXISTS aggregated_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_name VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    total_active_time INTEGER DEFAULT 0,
    total_afk_time INTEGER DEFAULT 0,
    session_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(app_name, date)
);

-- Create composite index for app_name and date lookups
CREATE INDEX IF NOT EXISTS idx_aggregated_stats_app_date ON aggregated_stats(app_name, date);
