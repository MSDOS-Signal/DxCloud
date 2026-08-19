-- ============================================================
-- 000001_init.sql  Phase 1 初始骨架表
-- 范围：IAM 基座（users/roles/permissions）+ 租户基座（organizations/projects）+ 系统设置
-- 约定：InnoDB / utf8mb4；id BIGINT 自增主键；created_at/updated_at 自动维护；
--       可软删表含 deleted_at；租户相关表含 org_id / project_id / owner_id（从 Phase 1 起就有，不后补）
-- ============================================================

CREATE TABLE IF NOT EXISTS users (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username      VARCHAR(64)     NOT NULL,
    email         VARCHAR(128)    NOT NULL,
    password_hash VARCHAR(255)    NOT NULL DEFAULT '',
    nickname      VARCHAR(64)     NOT NULL DEFAULT '',
    avatar_url    VARCHAR(512)    NOT NULL DEFAULT '',
    status        TINYINT         NOT NULL DEFAULT 1 COMMENT '1=active 2=disabled 3=locked',
    last_login_at DATETIME(3)     NULL,
    last_login_ip VARCHAR(64)     NOT NULL DEFAULT '',
    created_at    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at    DATETIME(3)     NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_username (username),
    UNIQUE KEY uk_users_email (email),
    KEY idx_users_status (status)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '平台用户';

CREATE TABLE IF NOT EXISTS roles (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code        VARCHAR(64)     NOT NULL,
    name        VARCHAR(64)     NOT NULL,
    description VARCHAR(255)    NOT NULL DEFAULT '',
    is_system   TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '1=系统内置角色，不可删除',
    scope       VARCHAR(16)     NOT NULL DEFAULT 'global' COMMENT 'global/org/project',
    created_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at  DATETIME(3)     NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_roles_code (code)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '角色';

CREATE TABLE IF NOT EXISTS permissions (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code        VARCHAR(64)     NOT NULL COMMENT '如 ecs:create',
    name        VARCHAR(64)     NOT NULL,
    module      VARCHAR(32)     NOT NULL DEFAULT '' COMMENT 'ecs/image/network/...',
    description VARCHAR(255)    NOT NULL DEFAULT '',
    created_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_permissions_code (code)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '权限点';

CREATE TABLE IF NOT EXISTS user_roles (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id    BIGINT UNSIGNED NOT NULL,
    role_id    BIGINT UNSIGNED NOT NULL,
    org_id     BIGINT UNSIGNED NULL COMMENT '组织级授权时填写',
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_user_roles (user_id, role_id, org_id),
    KEY idx_user_roles_org (org_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '用户-角色';

CREATE TABLE IF NOT EXISTS role_permissions (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    role_id       BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    created_at    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_role_permissions (role_id, permission_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '角色-权限';

CREATE TABLE IF NOT EXISTS login_logs (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id    BIGINT UNSIGNED NULL,
    ip         VARCHAR(64)     NOT NULL DEFAULT '',
    user_agent VARCHAR(512)    NOT NULL DEFAULT '',
    status     TINYINT         NOT NULL DEFAULT 0 COMMENT '1=成功 2=失败',
    message    VARCHAR(255)    NOT NULL DEFAULT '',
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_login_logs_user (user_id, created_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '登录日志';

CREATE TABLE IF NOT EXISTS organizations (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name       VARCHAR(128)    NOT NULL,
    code       VARCHAR(64)     NOT NULL,
    plan       VARCHAR(16)     NOT NULL DEFAULT 'free' COMMENT 'free/basic/pro',
    credit     DECIMAL(18, 4)  NOT NULL DEFAULT 0 COMMENT '虚拟余额',
    status     TINYINT         NOT NULL DEFAULT 1,
    created_by BIGINT UNSIGNED NULL,
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3)     NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_organizations_name (name),
    UNIQUE KEY uk_organizations_code (code)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '组织（租户）';

CREATE TABLE IF NOT EXISTS organization_members (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id     BIGINT UNSIGNED NOT NULL,
    user_id    BIGINT UNSIGNED NOT NULL,
    org_role   VARCHAR(16)     NOT NULL DEFAULT 'member' COMMENT 'owner/admin/member/viewer',
    status     TINYINT         NOT NULL DEFAULT 1,
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_org_members (org_id, user_id),
    KEY idx_org_members_user (user_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '组织成员';

CREATE TABLE IF NOT EXISTS projects (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id      BIGINT UNSIGNED NOT NULL,
    name        VARCHAR(128)    NOT NULL,
    code        VARCHAR(64)     NOT NULL,
    description VARCHAR(255)    NOT NULL DEFAULT '',
    status      TINYINT         NOT NULL DEFAULT 1,
    created_by  BIGINT UNSIGNED NULL,
    created_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at  DATETIME(3)     NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_projects_org_name (org_id, name),
    KEY idx_projects_org (org_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '项目';

CREATE TABLE IF NOT EXISTS project_environments (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    project_id BIGINT UNSIGNED NOT NULL,
    name       VARCHAR(16)     NOT NULL COMMENT 'development/testing/staging/production',
    seq        INT             NOT NULL DEFAULT 0,
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_project_envs (project_id, name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '项目环境';

CREATE TABLE IF NOT EXISTS system_settings (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `key`       VARCHAR(128)    NOT NULL,
    `value`     JSON            NULL,
    description VARCHAR(255)    NOT NULL DEFAULT '',
    updated_by  BIGINT UNSIGNED NULL,
    created_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_system_settings_key (`key`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '系统设置（配额默认值/安全策略/保留期等）';
