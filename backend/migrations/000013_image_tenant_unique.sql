-- 000013_image_tenant_unique.sql：镜像记录按租户隔离；去掉全局 repo/tag 唯一约束。
ALTER TABLE docker_images
    DROP INDEX uk_docker_images;

ALTER TABLE docker_images
    ADD KEY idx_docker_images_org_repo_tag (org_id, repo, tag);
