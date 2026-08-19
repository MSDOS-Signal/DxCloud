-- 000014_notification_link.sql
-- 通知中心增强：新增 link 字段（点击通知可跳转到对应功能页面）
ALTER TABLE notifications
    ADD COLUMN link VARCHAR(255) NULL COMMENT '前端跳转路径，如 /ecs/3 /images /apps' AFTER content;
