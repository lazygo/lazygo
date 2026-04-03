-- user

CREATE TABLE IF NOT EXISTS user (
  `uid` int unsigned NOT NULL AUTO_INCREMENT,
  `email` varchar(50) DEFAULT NULL COMMENT 'email',
  `mobile` varchar(20) DEFAULT NULL COMMENT '手机号',
  `password` varchar(100) NOT NULL DEFAULT '' COMMENT '密码',
  `ctime` int unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `mtime` int unsigned NOT NULL DEFAULT '0' COMMENT '修改时间',
  PRIMARY KEY (`uid`),
  UNIQUE KEY `uk_email` (`email`),
  UNIQUE KEY `uk_mobile` (`mobile`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户表';
