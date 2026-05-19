CREATE TABLE IF NOT EXISTS blog_categories (
  id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(160) NOT NULL,
  description VARCHAR(512) NOT NULL DEFAULT '',
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_blog_categories_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS blog_tags (
  id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
  slug VARCHAR(96) NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_blog_tags_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS blog_articles (
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
  UNIQUE KEY uk_blog_articles_slug (slug),
  KEY idx_blog_articles_author_id (author_id),
  KEY idx_blog_articles_category_id (category_id),
  KEY idx_blog_articles_status_published_at (status, published_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS blog_article_revisions (
  article_id BIGINT NOT NULL,
  version INT NOT NULL,
  title VARCHAR(160) NOT NULL,
  summary VARCHAR(512) NOT NULL DEFAULT '',
  content MEDIUMTEXT NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (article_id, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS blog_article_tags (
  article_id BIGINT NOT NULL,
  tag_id BIGINT NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (article_id, tag_id),
  KEY idx_blog_article_tags_tag_id (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
