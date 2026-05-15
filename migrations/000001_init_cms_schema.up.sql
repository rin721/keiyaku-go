CREATE TABLE IF NOT EXISTS users (
  id BIGINT NOT NULL,
  username VARCHAR(32) NOT NULL,
  email VARCHAR(254) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  roles_json JSON NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_username (username),
  UNIQUE KEY uk_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS roles (
  id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
  description VARCHAR(255) NOT NULL DEFAULT '',
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_roles_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS permissions (
  id BIGINT NOT NULL,
  code VARCHAR(128) NOT NULL,
  description VARCHAR(255) NOT NULL DEFAULT '',
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_permissions_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_roles (
  user_id BIGINT NOT NULL,
  role_id BIGINT NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (user_id, role_id),
  KEY idx_user_roles_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS role_permissions (
  role_id BIGINT NOT NULL,
  permission_id BIGINT NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (role_id, permission_id),
  KEY idx_role_permissions_permission_id (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS categories (
  id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(160) NOT NULL,
  description VARCHAR(512) NOT NULL DEFAULT '',
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_categories_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tags (
  id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
  slug VARCHAR(96) NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_tags_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS articles (
  id BIGINT NOT NULL,
  author_id BIGINT NOT NULL,
  category_id BIGINT NOT NULL DEFAULT 0,
  title VARCHAR(160) NOT NULL,
  slug VARCHAR(180) NOT NULL,
  summary VARCHAR(512) NOT NULL DEFAULT '',
  content MEDIUMTEXT NOT NULL,
  status VARCHAR(32) NOT NULL,
  tags_json JSON NOT NULL,
  published_at TIMESTAMP(6) NULL,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_articles_slug (slug),
  KEY idx_articles_author_id (author_id),
  KEY idx_articles_category_id (category_id),
  KEY idx_articles_status_published_at (status, published_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS article_tags (
  article_id BIGINT NOT NULL,
  tag_id BIGINT NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (article_id, tag_id),
  KEY idx_article_tags_tag_id (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS comments (
  id BIGINT NOT NULL,
  article_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  parent_id BIGINT NOT NULL DEFAULT 0,
  content TEXT NOT NULL,
  status VARCHAR(32) NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_comments_article_id (article_id),
  KEY idx_comments_user_id (user_id),
  KEY idx_comments_parent_id (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS casbin_rule (
  id BIGINT AUTO_INCREMENT NOT NULL,
  ptype VARCHAR(100) DEFAULT NULL,
  v0 VARCHAR(100) DEFAULT NULL,
  v1 VARCHAR(100) DEFAULT NULL,
  v2 VARCHAR(100) DEFAULT NULL,
  v3 VARCHAR(100) DEFAULT NULL,
  v4 VARCHAR(100) DEFAULT NULL,
  v5 VARCHAR(100) DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
