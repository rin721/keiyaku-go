ALTER TABLE plugin_instances
  ADD COLUMN health_status VARCHAR(32) NOT NULL DEFAULT 'unknown' AFTER status,
  ADD COLUMN last_checked_at TIMESTAMP(6) NULL AFTER lease_expires_at,
  ADD COLUMN consecutive_failures INT NOT NULL DEFAULT 0 AFTER last_checked_at,
  ADD COLUMN last_error_at TIMESTAMP(6) NULL AFTER last_error;

CREATE INDEX idx_plugin_instances_health
  ON plugin_instances (plugin_key, manifest_hash, status, health_status, lease_expires_at);

CREATE TABLE IF NOT EXISTS plugin_audit_events (
  id BIGINT AUTO_INCREMENT NOT NULL,
  plugin_key VARCHAR(64) NOT NULL,
  instance_id VARCHAR(64) NOT NULL DEFAULT '',
  action VARCHAR(64) NOT NULL,
  message VARCHAR(255) NOT NULL DEFAULT '',
  metadata_json JSON NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_plugin_audit_plugin_created (plugin_key, created_at),
  KEY idx_plugin_audit_action (action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
