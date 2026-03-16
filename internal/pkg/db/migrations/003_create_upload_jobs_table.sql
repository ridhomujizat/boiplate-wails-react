CREATE TABLE IF NOT EXISTS upload_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    filename VARCHAR(512),
    content_type VARCHAR(255),
    file_size INTEGER DEFAULT 0,
    status VARCHAR(64) NOT NULL DEFAULT 'pending',
    attempt_count INTEGER DEFAULT 0,
    next_attempt_at DATETIME,
    last_attempt_at DATETIME,
    last_error TEXT,
    last_http_status INTEGER DEFAULT 0,
    upload_id VARCHAR(255),
    signed_url TEXT,
    signed_url_expires_at DATETIME,
    gcs_uploaded_at DATETIME,
    confirmed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_upload_jobs_session_id ON upload_jobs(session_id);
CREATE INDEX IF NOT EXISTS idx_upload_jobs_status ON upload_jobs(status);
CREATE INDEX IF NOT EXISTS idx_upload_jobs_next_attempt_at ON upload_jobs(next_attempt_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_upload_jobs_session_file ON upload_jobs(session_id, file_path);
