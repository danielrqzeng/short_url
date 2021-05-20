// gen by iyfiysi at 2021 May 19

package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"iyfiysi.com/short_url/internal/app/server"
	"iyfiysi.com/short_url/internal/pkg/conf"
	"iyfiysi.com/short_url/internal/pkg/db"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/model"
	"iyfiysi.com/short_url/internal/pkg/trace"
	"iyfiysi.com/short_url/internal/pkg/utils"
	"os"
	"strings"
)

//定义一个全局变量的命令行接收参数
var (
	etcdServerFlag = flag.String("etcd", "http://127.0.0.1:2379", `etcd server,split with "," if more than one etcd server`)
	confKeyFlag    = flag.String("conf_key", "/short_url/config/app.yaml", `etcd conf key`)
	versionFlag    = flag.Bool("version", false, "print the current version")
)

// Variables set at build time
var (
	version = "v1.0.0"
	commit  = "unknown"
	date    = "unknown"
)

func initAll() {
	conf.Init()
	err := conf.InitRemoteConfig(
		strings.Split(*etcdServerFlag, ","),
		*confKeyFlag,
		func() {
			fmt.Println(viper.GetString("version"))
		})
	if err != nil {
		panic(err)
	}
	//做各个部件的初始化
	logger.Init()
	utils.Init()
	trace.Init()
	db.Init()
	model.Init()
}

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Version %v, commit %v, built at %v\n", version, commit, date)
		os.Exit(0)
	}
	defer utils.DeferWhenCoreDump()

	initAll()

	if err := server.App().Run(); err != nil {
		fmt.Println(err)
	}
}
