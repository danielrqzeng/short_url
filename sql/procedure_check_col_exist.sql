set @db_name 	= 'none';
SELECT database() into @db_name;
set @tbl_name	= 'none';
set @col_name	= 'none';
-- the incr path
-- NOTE: param @in_col_name can not bigger than 255 char
-- 		 param @col_define 	can not bigger than 255 char
DELIMITER $$
DROP PROCEDURE IF EXISTS `addColIfNotExist` $$
CREATE PROCEDURE addColIfNotExist(
	in in_col_name char(255),
	in col_define char(255) charset 'utf8'
--	in col_val_insert char(100)
)
BEGIN
	declare col_num int;
	select count(*) into col_num FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@db_name AND table_name=@tbl_name AND COLUMN_NAME=in_col_name;
	-- if col_num = 1 then
	-- 	select concat('errmsg : ',@db_name,'.',@tbl_name,'.',in_col_name,' exist') as 'errno  : fail';
	-- else
	if col_num = 0 then
		set @define	= col_define;
		PREPARE define FROM @define;
		execute define;
		DEALLOCATE PREPARE define;
		select concat('msg  : column ',@db_name,'.',@tbl_name,'.',in_col_name,' new success') as 'rslt : success';
	end if;
END$$

DELIMITER ;

