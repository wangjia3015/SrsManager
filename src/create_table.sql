use srs_manager;

CREATE TABLE `room` (
      `id` bigint(20) NOT NULL AUTO_INCREMENT,
      `user` varchar(255) NOT NULL,
      `desc` varchar(255) NOT NULL,
      `streamname` varchar(255) NOT NULL,
      `expiration` int(11) NOT NULL,
      `status` int(11) NOT NULL,
      `publishid` int(11) DEFAULT '-1',
      `publishhost` varchar(20) DEFAULT '',
      `lastupdatetime` int(11) NOT NULL,
      `createtime` int(11) NOT NULL,
      PRIMARY KEY (`id`),
      UNIQUE KEY `name` (`streamname`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `srs_server` (
      `id` bigint(20) NOT NULL AUTO_INCREMENT,
      `host` varchar(255) NOT NULL,
      `desc` varchar(255) DEFAULT '',
      `type` int(11) NOT NULL,
      `status` int(11) NOT NULL,
      PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8
