-- 認証に必要なテーブル定義
-- User table
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'アプリ内ユーザーの内部ID',
  email VARCHAR(255) NOT NULL COMMENT 'メールアドレス（uk_users_emailで一意）',
  display_name VARCHAR(255) NOT NULL COMMENT '表示名（プロフィール名）',
  picture_url VARCHAR(512) NULL COMMENT 'プロフィール画像URL（任意）',
  email_verified TINYINT(1) NOT NULL DEFAULT 0 COMMENT 'メール認証済みフラグ（0/1）',
  status ENUM('active','suspended','deleted') NOT NULL DEFAULT 'active' COMMENT 'ユーザー状態（active/suspended/deleted）',
  last_login_at DATETIME(6) NULL COMMENT '最終ログイン時刻（認証成功時に更新）',
  created_at DATETIME(6) NOT NULL COMMENT '作成時刻',
  updated_at DATETIME(6) NOT NULL COMMENT '更新時刻',
  deleted_at DATETIME(6) NULL COMMENT '論理削除時刻（NULLなら有効）',
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='アプリ内ユーザー（基本プロフィール）';

-- UserIdentity (OIDC連携)
CREATE TABLE user_identities (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID連携の内部ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT 'users.id への参照',
  provider VARCHAR(64) NOT NULL COMMENT 'IDプロバイダ名（例: cognito, google, github, keycloak）',
  issuer VARCHAR(255) NOT NULL COMMENT 'OIDC Issuer（例: https://.../realms/xxx）',
  subject VARCHAR(255) NOT NULL COMMENT 'OIDC subject（発行者内で一意なユーザーID）',
  email_at_provider VARCHAR(255) NULL COMMENT 'プロバイダ側で確認されたメール（任意）',
  connected_at DATETIME(6) NOT NULL COMMENT '初回リンク時刻（アカウント連携開始）',
  last_login_at DATETIME(6) NULL COMMENT 'このプロバイダ経由の最終ログイン時刻',
  created_at DATETIME(6) NOT NULL COMMENT '作成時刻',
  updated_at DATETIME(6) NOT NULL COMMENT '更新時刻',
  PRIMARY KEY (id),
  UNIQUE KEY uk_oidc_iss_sub (issuer, subject),
  KEY idx_user_id (user_id),
  CONSTRAINT fk_user_identities_user FOREIGN KEY (user_id) REFERENCES users(id)
    ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='外部IDプロバイダ（OIDC）とユーザーのひも付け';

-- Role / UserRole（最小RBAC）
CREATE TABLE roles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ロールの内部ID',
  name VARCHAR(64) NOT NULL COMMENT 'ロール名（例: admin, user）',
  description VARCHAR(255) NULL COMMENT 'ロールの説明（任意）',
  created_at DATETIME(6) NOT NULL COMMENT '作成時刻',
  updated_at DATETIME(6) NOT NULL COMMENT '更新時刻',
  PRIMARY KEY (id),
  UNIQUE KEY uk_roles_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='ロール定義（最小RBAC）';

CREATE TABLE user_roles (
  user_id BIGINT UNSIGNED NOT NULL COMMENT 'users.id への参照',
  role_id BIGINT UNSIGNED NOT NULL COMMENT 'roles.id への参照',
  created_at DATETIME(6) NOT NULL COMMENT '付与時刻',
  PRIMARY KEY (user_id, role_id),
  CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users(id)
    ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles(id)
    ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='ユーザーとロールの付与関係（多対多）';

-- セッション（BFF+Cookie採用時のみ）
CREATE TABLE sessions (
  id CHAR(36) NOT NULL COMMENT 'セッションID（UUID）',
  user_id BIGINT UNSIGNED NOT NULL COMMENT 'users.id への参照',
  user_agent VARCHAR(255) NULL COMMENT 'ユーザーエージェント（任意）',
  ip_address VARCHAR(64) NULL COMMENT 'アクセス元IP（任意）',
  expires_at DATETIME(6) NOT NULL COMMENT '有効期限',
  revoked_at DATETIME(6) NULL COMMENT '失効時刻（NULLなら有効）',
  created_at DATETIME(6) NOT NULL COMMENT '作成時刻',
  PRIMARY KEY (id),
  KEY idx_sessions_user (user_id),
  CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id)
    ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='BFF+Cookie採用時のサーバサイドセッション';

-- Sample table
CREATE TABLE samples (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  content VARCHAR(255) NOT NULL,
  count INT UNSIGNED NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


-- Article table
CREATE TABLE articles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  content VARCHAR(255) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

