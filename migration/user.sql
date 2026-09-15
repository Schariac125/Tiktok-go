-- 用户表
-- 对应 proto: app/user/rpc/user.proto -> User 消息

CREATE TABLE IF NOT EXISTS `user` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID 编号',
    `username`    VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '用户名',
    `password`    VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '密码（加密存储）',
    `avatar_url`  VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '头像链接地址',
    `mfa_secret`  VARCHAR(255)   NULL DEFAULT NULL COMMENT 'mfa 密钥（TOTP secret，未绑定时为 NULL）',
    `mfa_enabled` TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '是否已启用 mfa：0-未启用，1-已启用',
    `create_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_at`   DATETIME        NULL DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='用户表';
