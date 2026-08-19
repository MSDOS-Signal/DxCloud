-- ============================================================
-- 000002_audit_logs.sql  Phase 2：审计日志表
-- ============================================================

CREATE TABLE IF NOT EXISTS audit_logs (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id        BIGINT UNSIGNED NULL,
    user_id       BIGINT UNSIGNED NULL,
    action        VARCHAR(64)  NOT NULL,
    resource_type VARCHAR(32)  NOT NULL DEFAULT '',
    resource_id   VARCHAR(64)  NOT NULL DEFAULT '',
    ip            VARCHAR(64)  NOT NULL DEFAULT '',
    request_id    VARCHAR(64)  NOT NULL DEFAULT '',
    detail        TEXT         NULL,
    status        TINYINT      NOT NULL DEFAULT 1 COMMENT '1=成功 2=拒绝',
    created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_audit_org (org_id, created_at),
    KEY idx_audit_user (user_id, created_at),
    KEY idx_audit_resource (resource_type, resource_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '审计日志';
