// gen by iyfiysi at 2021 May 19
package service

// Init 初始化，程序启动时候调用
func Init() {
	err := IDMgr().IDInit()
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
