#!/bin/bash

mysql_host='127.0.0.1'
mysql_port=3306
mysql_user='root'
mysql_pass='root'

# db databases name
db_array=(
short_url
)


function create_table(){
    sql_file=$1
	for i in ${db_array[@]}; do
		if [ -z "$mysql_pass" ];then
    		mysql -h$mysql_host -P$mysql_port -u$mysql_user $i <$sql_file
		fi
		if [ -n "$mysql_pass" ];then
    		mysql -h$mysql_host -P$mysql_port -u$mysql_user -p$mysql_pass $i <$sql_file
		fi
	done
}

for file in `ls ./*.sql|grep -v procedure_check_col_exist`; do
	echo $file
    create_table $file
done


