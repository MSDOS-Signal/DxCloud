-- 000012_image_pull_logs.sql：镜像拉取实时进度日志
ALTER TABLE docker_images
    ADD COLUMN pull_log MEDIUMTEXT NULL COMMENT '镜像拉取实时日志' AFTER pull_error;
