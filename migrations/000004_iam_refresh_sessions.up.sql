CREATE TABLE IF NOT EXISTS iam_refresh_sessions (
  id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  refresh_token_id VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL,
  replaced_by_session_id BIGINT NULL,
  expires_at TIMESTAMP(6) NOT NULL,
  revoked_at TIMESTAMP(6) NULL,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_iam_refresh_sessions_token_id (refresh_token_id),
  KEY idx_iam_refresh_sessions_user_status (user_id, status),
  KEY idx_iam_refresh_sessions_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
