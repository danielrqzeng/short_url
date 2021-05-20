// gen by iyfiysi at 2021 May 19

package trace

import (
	"github.com/spf13/viper"
)

// Init 初始化，程序启动时候调用
func Init() {
	enable := viper.GetBool("jaeger.enable")
	if !enable {
		return
	}

	jaegerAddrs := viper.GetStringSlice("jaeger.jaegerServer")
	if len(jaegerAddrs) == 0 {
		panic("len(jaeger.jaegerServer)=0")
	}
	//全局最终实例初始化
	err := InitTracer(jaegerAddrs[0], "short_url")
	if err != nil {
		panic(err)
	}
}

// OnConfigChange 配置变更的通知
func OnConfigChange() {

}

// OnShutDown 服务被终止的通知
func OnShutDown() {

}
