CREATE DATABASE IF NOT EXISTS `test` default character set utf8mb4 collate utf8mb4_unicode_ci;
use `test`;

SET NAMES utf8mb4;


-- ----------------------------
-- Table structure for command
-- ----------------------------
DROP TABLE IF EXISTS `command`;
CREATE TABLE `command`
(
    `id`          bigint                                                       NOT NULL AUTO_INCREMENT,
    `sender_id`   bigint                                                       NOT NULL COMMENT '用户ID',
    `command`     text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci        NOT NULL COMMENT '账单金额',
    `create_time` datetime                                                     NOT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '账单记录生成时间',
    `args`        text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci        NOT NULL,
    `status`      varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '1' COMMENT '‘1’: 有效；‘2’:无效',
    PRIMARY KEY (`id`),
    KEY `idx_create_time` (`create_time`) USING BTREE
) ENGINE = InnoDB
  AUTO_INCREMENT = 35
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- ----------------------------
-- Table structure for menu
-- ----------------------------
DROP TABLE IF EXISTS `menu`;
CREATE TABLE `menu`
(
    `id`            bigint                                                        NOT NULL AUTO_INCREMENT,
    `name`          varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '食物名称',
    `code`          varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '食物编码',
    `price`         int                                                           NOT NULL COMMENT '食物价格',
    `currency_type` int                                                           NOT NULL COMMENT '货币类型',
    `supplier`      varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '供应商',
    PRIMARY KEY (`id`),
    UNIQUE KEY `code` (`code`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 15
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`
(
    `id`         bigint                                                        NOT NULL COMMENT '用户ID',
    `name`       varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '用户名称',
    `status`     varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL DEFAULT '1' COMMENT '用户状态 ‘1’:正常；’2’:不可用',
    `timezone`   int                                                           NOT NULL DEFAULT '0' COMMENT '时区',
    `group_name` varchar(255) COLLATE utf8mb4_general_ci                       NOT NULL DEFAULT '' COMMENT '群名',
    `language`   varchar(8) COLLATE utf8mb4_general_ci                         NOT NULL DEFAULT 'zh' COMMENT '语言',
    PRIMARY KEY (`id`),
    UNIQUE KEY `name` (`name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- ----------------------------
-- Table structure for wallet
-- ----------------------------
DROP TABLE IF EXISTS `wallet`;
CREATE TABLE `wallet`
(
    `id`                  bigint NOT NULL AUTO_INCREMENT,
    `user_id`             bigint NOT NULL,
    `type`                int    NOT NULL,
    `account_numerator`   bigint NOT NULL,
    `account_denominator` bigint NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `wallet_user_id_type_uindex` (`user_id`, `type`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 15
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- ----------------------------
-- Table structure for wallet_log
-- ----------------------------
DROP TABLE IF EXISTS `wallet_log`;
CREATE TABLE `wallet_log`
(
    `id`                 bigint   NOT NULL AUTO_INCREMENT,
    `command_id`         bigint   NOT NULL,
    `user_id`            bigint   NOT NULL,
    `type`               int      NOT NULL,
    `before_numerator`   bigint   NOT NULL,
    `before_denominator` bigint   NOT NULL,
    `inc`                double   NOT NULL,
    `after_numerator`    bigint   NOT NULL,
    `after_denominator`  bigint   NOT NULL,
    `create_time`        datetime NOT NULL,
    PRIMARY KEY (`id`),
    KEY `command_id_user_id_index` (`command_id`, `user_id`) USING BTREE
) ENGINE = InnoDB
  AUTO_INCREMENT = 93
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;
