package model

//package main
//
//import (
//	"fmt"
//	"github.com/gohouse/converter"
//)
//
//func main() {
////StructNameToHump:true为了让表名也是大写，比如对表名为t_kv_info,默认生成的结构体也为t_kv_info,添加了StructNameToHump:true之后生成的结构体为TKvInfo
//	err := converter.NewTable2Struct().
//		Config(&converter.T2tConfig{SeperatFile: true,StructNameToHump:true}).
//		SavePath("model_test.go").
//		Dsn("root:root@tcp(127.0.0.1:3306)/short_url?charset=utf8").
//		EnableJsonTag(true).
//		PackageName("model").
//		RealNameMethod("TableName").
//		TagKey("gorose").
//		Run()
//	fmt.Println(err)
//}

