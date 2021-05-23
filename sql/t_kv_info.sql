/*!40101 SET NAMES utf8 */;

/*!40101 SET SQL_MODE=''*/;

/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
/*Table structure for table `t_kv_info` */


CREATE TABLE IF NOT EXISTS `t_kv_info` (
	`key_info`		varchar(255) 		        NOT NULL COMMENT 'key_info',
	`val_str_info`		blob 				        NOT NULL COMMENT 'val,字符串格式',
	`val_int_info`		bigint(20) unsigned NOT NULL COMMENT 'val,整数格式',
	`desc_info`       blob				        NOT NULL COMMENT 'desc',
	`create_ts`	      int(10) unsigned	  NOT NULL DEFAULT 0 COMMENT '创建时间',
	`version`	        int(10) unsigned 	  NOT NULL DEFAULT 0 COMMENT '数据版本,cas update辅佐作用',
	PRIMARY KEY (`key_info`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

source procedure_check_col_exist.sql
set @db_name 	= 'none';
SELECT database() into @db_name;
set @tbl_name	= 't_kv_info';
set @col_name	= 'none';


set @col_name 		= 'last_update_ts';
set @field_define 	= concat('ALTER TABLE `',@tbl_name,'` ADD COLUMN `',@col_name,'` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT \'上次更新时间戳\'');
call addColIfNotExist(@col_name,@field_define);


DROP PROCEDURE IF EXISTS `addColIfNotExist`;

