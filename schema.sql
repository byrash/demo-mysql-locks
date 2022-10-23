set global innodb_print_all_deadlocks = 1;
SET global general_log = on;
SET global general_log_file='/var/log/mysql/mysql.log';
SET global log_output = 'file';

CREATE TABLE `parent` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;


CREATE TABLE `child` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(50) DEFAULT NULL,
  `parent_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `parent_id` (`parent_id`),
  CONSTRAINT `child_ibfk_1` FOREIGN KEY (`parent_id`) REFERENCES `parent` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
