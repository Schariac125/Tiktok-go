-- 社交模块的关注关系表
-- 对应 proto: api/social/social.proto
--
-- 好友关系不单独存表：互相关注即为好友，用自连接查出来即可，
-- 省去一张表也就省去了两边同步的麻烦。

CREATE TABLE IF NOT EXISTS `user_relation` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID 编号',
    `user_id`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关注者 id',
    `to_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '被关注者 id',
    `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '关注时间',
    PRIMARY KEY (`id`),
    -- 同一对关系只允许一条，防重复关注；也是「我关注的人」列表的驱动索引
    UNIQUE KEY `uk_user_to` (`user_id`, `to_user_id`),
    -- 「关注我的人」（粉丝列表）
    KEY `idx_to_user` (`to_user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='用户关注关系表';
