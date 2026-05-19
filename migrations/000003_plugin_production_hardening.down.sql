DROP TABLE IF EXISTS plugin_audit_events;

DROP INDEX idx_plugin_instances_health ON plugin_instances;

ALTER TABLE plugin_instances
  DROP COLUMN last_error_at,
  DROP COLUMN consecutive_failures,
  DROP COLUMN last_checked_at,
  DROP COLUMN health_status;
