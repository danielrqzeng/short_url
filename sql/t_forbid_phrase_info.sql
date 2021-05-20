/*!40101 SET NAMES utf8 */;

/*!40101 SET SQL_MODE=''*/;

/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
/*Table structure for table `t_forbid_phrase_info` */

-- 禁止短语表
CREATE TABLE IF NOT EXISTS `t_forbid_phrase_info` (
  `id`            int(10) unsigned    NOT NULL AUTO_INCREMENT COMMENT 'id',
	`phrase_type`	  int(10) unsigned	  NOT NULL DEFAULT 0 COMMENT '短语类型，0:none,1:短语,2:正则表达式',
	`phrase`		    varchar(64) 		    NOT NULL COMMENT '短语|正则式',
	`create_ts`	    int(10) unsigned	  NOT NULL DEFAULT 0 COMMENT '创建时间',
	`version`	      int(10) unsigned 	  NOT NULL DEFAULT 0 COMMENT '数据版本,cas update辅佐作用',

	unique key phrase_idx(`phrase`),
	PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

source procedure_check_col_exist.sql
set @db_name 	= 'none';
SELECT database() into @db_name;
set @tbl_name	= 't_forbid_phrase_info';
set @col_name	= 'none';


set @col_name 		= 'last_update_ts';
set @field_define 	= concat('ALTER TABLE `',@tbl_name,'` ADD COLUMN `',@col_name,'` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT \'上次更新时间戳\'');
call addColIfNotExist(@col_name,@field_define);


DROP PROCEDURE IF EXISTS `addColIfNotExist`;

