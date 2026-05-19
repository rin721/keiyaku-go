DROP TABLE IF EXISTS plugin_route_claims;
DROP TABLE IF EXISTS plugin_signature_nonces;
DROP TABLE IF EXISTS plugin_audit_events;
DROP TABLE IF EXISTS plugin_routes;
DROP TABLE IF EXISTS plugin_instances;
DROP TABLE IF EXISTS plugin_services;

CREATE TABLE plugin_services (
  id BIGINT AUTO_INCREMENT NOT NULL,
  plugin_key VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  protocol VARCHAR(32) NOT NULL,
  current_manifest_hash CHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  metadata_json JSON NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  disabled_at TIMESTAMP(6) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_services_key (plugin_key),
  KEY idx_plugin_services_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE plugin_instances (
  id BIGINT AUTO_INCREMENT NOT NULL,
  plugin_key VARCHAR(64) NOT NULL,
  instance_id VARCHAR(64) NOT NULL,
  version VARCHAR(64) NOT NULL,
  base_url VARCHAR(512) NOT NULL,
  health_path VARCHAR(255) NOT NULL,
  manifest_hash CHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  health_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
  last_seen_at TIMESTAMP(6) NOT NULL,
  lease_expires_at TIMESTAMP(6) NOT NULL,
  last_checked_at TIMESTAMP(6) NULL,
  consecutive_failures INT NOT NULL DEFAULT 0,
  last_error VARCHAR(512) NOT NULL DEFAULT '',
  last_error_at TIMESTAMP(6) NULL,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_instances_key_instance (plugin_key, instance_id),
  KEY idx_plugin_instances_routable (plugin_key, manifest_hash, status, health_status, lease_expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE plugin_routes (
  id BIGINT AUTO_INCREMENT NOT NULL,
  plugin_key VARCHAR(64) NOT NULL,
  manifest_hash CHAR(64) NOT NULL,
  route_id VARCHAR(64) NOT NULL,
  method VARCHAR(16) NOT NULL,
  match_type VARCHAR(16) NOT NULL,
  gateway_path VARCHAR(255) NOT NULL,
  upstream_path VARCHAR(255) NOT NULL,
  auth_policy VARCHAR(32) NOT NULL,
  timeout VARCHAR(32) NOT NULL,
  forward_auth_header TINYINT(1) NOT NULL DEFAULT 0,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  metadata_json JSON NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_routes_route_id (plugin_key, manifest_hash, route_id),
  UNIQUE KEY uk_plugin_routes_gateway (plugin_key, manifest_hash, method, match_type, gateway_path),
  KEY idx_plugin_routes_lookup (plugin_key, manifest_hash, enabled, method, gateway_path)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE plugin_audit_events (
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

CREATE TABLE plugin_signature_nonces (
  id BIGINT AUTO_INCREMENT NOT NULL,
  plugin_key VARCHAR(64) NOT NULL,
  nonce VARCHAR(128) NOT NULL,
  expires_at TIMESTAMP(6) NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_signature_nonces_key (plugin_key, nonce),
  KEY idx_plugin_signature_nonces_expiry (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
