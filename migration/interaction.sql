-- 互动模块的两张表：评论、点赞
-- 对应 proto: api/interaction/interaction.proto

-- 评论表
CREATE TABLE IF NOT EXISTS `comment` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID 编号',
    `user_id`     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '发布者用户 id',
    `video_id`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '视频 id',
    `parent_id`   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父评论 id，0 表示直接评论视频',
    `like_count`  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞数',
    `child_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '子评论数',
    `content`     VARCHAR(1024)   NOT NULL DEFAULT '' COMMENT '评论文本',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`  DATETIME        NULL DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    -- 查某个视频的一级评论
    KEY `idx_video_parent` (`video_id`, `parent_id`),
    -- 查某条评论下的回复
    KEY `idx_parent` (`parent_id`),
    -- 查某个用户发过的评论
    KEY `idx_user` (`user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='评论表';

-- 点赞表：视频与评论共用一张表，用 target_type 区分
-- 表名不叫 like，那是 SQL 保留字
CREATE TABLE IF NOT EXISTS `user_like` (
    `id`          BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT COMMENT 'ID 编号',
    `user_id`     BIGINT UNSIGNED  NOT NULL DEFAULT 0 COMMENT '点赞者 id',
    `target_type` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞目标类型：1-视频，2-评论',
    `target_id`   BIGINT UNSIGNED  NOT NULL DEFAULT 0 COMMENT '目标 id（视频 id 或评论 id）',
    `created_at`  DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    -- 同一用户对同一目标只能有一条记录，也用于判断「是否已点赞」
    UNIQUE KEY `uk_user_target` (`user_id`, `target_type`, `target_id`),
    -- 统计某个目标被点赞的总数
    KEY `idx_target` (`target_type`, `target_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='点赞表';
