-- ============================================================
-- 000011_settings.sql  Phase 15：系统设置（区域 / 镜像加速源）
-- ============================================================

CREATE TABLE IF NOT EXISTS system_settings (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `key`       VARCHAR(128) NOT NULL COMMENT '设置键（region / registry_mirror）',
    value       JSON NOT NULL COMMENT 'JSON 标量（字符串）',
    description VARCHAR(255) NOT NULL DEFAULT '',
    updated_by  BIGINT UNSIGNED NULL,
    created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_system_settings_key (`key`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '系统设置（key 唯一）';

-- 默认值：中国大陆区域 + 公共加速源（可在 设置 → 区域与镜像源 中修改）
INSERT INTO system_settings (`key`, value, description)
VALUES ('region', '"cn"', '区域：cn=中国大陆 / global=非中国大陆')
ON DUPLICATE KEY UPDATE `key` = `key`;

INSERT INTO system_settings (`key`, value, description)
VALUES ('registry_mirror', '"docker.1ms.run"', '中国大陆区域拉取官方镜像使用的加速源域名')
ON DUPLICATE KEY UPDATE `key` = `key`;
