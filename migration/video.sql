-- 视频表
-- 对应 proto: api/video/video.proto -> Video 消息

CREATE TABLE IF NOT EXISTS `video` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID 编号',
    `user_id`       BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '视频作者 id',
    `video_url`     VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '视频链接',
    `cover_url`     VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '封面链接',
    `title`         VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '视频标题',
    `description`   VARCHAR(1024)   NOT NULL DEFAULT '' COMMENT '视频描述',
    `visit_count`   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '观看量',
    `like_count`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞量',
    `comment_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '评论数',
    `created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_at`     DATETIME        NULL DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    -- 查看某个用户投稿的视频
    KEY `idx_user_id` (`user_id`),
    -- 首页视频流按时间倒序取，搜索按时间区间过滤
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='视频表';
