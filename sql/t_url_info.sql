/*!40101 SET NAMES utf8 */;

/*!40101 SET SQL_MODE=''*/;

/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
/*Table structure for table `t_url_info` */


CREATE TABLE IF NOT EXISTS `t_url_info` (
  `id`          bigint(20) unsigned   NOT NULL AUTO_INCREMENT COMMENT 'id',
	`short_code`	varchar(64) 		      NOT NULL default '' COMMENT '短码',
	`short_url`	  varchar(128) 		      NOT NULL default '' COMMENT '短网址',
	`raw_url`	    varchar(1024) 		    NOT NULL default '' COMMENT '原始网址',
	`url_type`	  int(10) unsigned	    NOT NULL COMMENT '类型，0:none,1:incr id,2:phrase',
	`status`	    int(10) unsigned	    NOT NULL COMMENT '状态，0:none,1:enable,2:ban,3:notexist',
	`ban_at`	    int(10) unsigned	    NOT NULL COMMENT '被禁时间戳',
	`unban_at`	  int(10) unsigned	    NOT NULL COMMENT '解禁时间戳',
	`ban_cuz`		  varchar(255) 		      NOT NULL COMMENT '禁止原因',
	`ban_by`		  varchar(255) 		      NOT NULL COMMENT '被谁禁止，sys为系统自动禁止',
	`redirect_time`	    int(10) unsigned	NOT NULL COMMENT '重定向次数',
	`last_redirect_ts`	int(10) unsigned	NOT NULL COMMENT '最近一次重定向的时间戳',
	`create_ts`		int(10) unsigned	NOT NULL DEFAULT 0 COMMENT '创建时间',
	`version`			int(10) unsigned 	NOT NULL DEFAULT 0 COMMENT '数据版本,cas update辅佐作用',

  index url_type_idx(`url_type`),
	unique key short_code_idx(`short_code`),
	PRIMARY KEY pk(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

source procedure_check_col_exist.sql
set @db_name 	= 'none';
SELECT database() into @db_name;
set @tbl_name	= 't_url_info';
set @col_name	= 'none';


set @col_name 		= 'last_update_ts';
set @field_define 	= concat('ALTER TABLE `',@tbl_name,'` ADD COLUMN `',@col_name,'` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT \'上次更新时间戳\'');
call addColIfNotExist(@col_name,@field_define);



DROP PROCEDURE IF EXISTS `addColIfNotExist`;

