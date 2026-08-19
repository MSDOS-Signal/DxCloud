-- ============================================================
-- 000010_security.sql  Phase 11：密钥托管 / 安全扫描报告
-- ============================================================

CREATE TABLE IF NOT EXISTS secrets (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id       BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=全局/默认空间，其余=组织内',
    name         VARCHAR(128) NOT NULL,
    value_cipher TEXT NOT NULL COMMENT 'AES-256-GCM 密文（base64，密钥由 JWT_SECRET 派生）',
    created_by   BIGINT UNSIGNED NULL,
    created_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at   DATETIME(3) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_secrets_org_name (org_id, name, deleted_at),
    KEY idx_secrets_org (org_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '密钥托管（加密存储）';

CREATE TABLE IF NOT EXISTS security_reports (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    kind          VARCHAR(32) NOT NULL COMMENT 'baseline/image/secret',
    score         INT NOT NULL DEFAULT 100,
    finding_count INT NOT NULL DEFAULT 0,
    summary       TEXT NULL COMMENT '发现项 JSON 数组',
    created_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_security_reports_kind (kind, id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '安全扫描报告';
