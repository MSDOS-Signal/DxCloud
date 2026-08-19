-- ============================================================
-- 000007_webhooks.sql  Phase 8：Git Webhook
-- ============================================================

CREATE TABLE IF NOT EXISTS webhooks (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    project_id    BIGINT UNSIGNED NULL,
    pipeline_id   BIGINT UNSIGNED NOT NULL,
    provider      VARCHAR(16)  NOT NULL COMMENT 'github/gitlab/gitee',
    secret_enc    VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'AES-GCM 加密的签名密钥',
    branch_filter VARCHAR(128) NOT NULL DEFAULT '' COMMENT '如 main 或 release/*（空=全部）',
    events        VARCHAR(64)  NOT NULL DEFAULT 'push',
    status        TINYINT      NOT NULL DEFAULT 1,
    hook_code     VARCHAR(32)  NOT NULL COMMENT 'URL 随机码',
    created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at    DATETIME(3)  NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_webhooks_code (hook_code),
    KEY idx_webhooks_pipeline (pipeline_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'Git Webhook 注册';
