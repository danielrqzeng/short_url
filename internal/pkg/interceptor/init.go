// gen by iyfiysi at 2021 May 19

package interceptor

import (
	httpInterceptor "iyfiysi.com/short_url/internal/pkg/interceptor/http"
)

// Init 初始化，程序启动时候调用
func Init() {
	httpInterceptor.Init()
}

// OnConfigChange 配置变更的通知
func OnConfigChange() {

}

// OnShutDown 服务被终止的通知
func OnShutDown() {

}
