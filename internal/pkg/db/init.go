// gen by iyfiysi at 2021 May 19

package db

// Init 初始化，程序启动时候调用
func Init() {
	RedisInit()
	MysqlInit()
}

// OnConfigChange 配置变更的通知
func OnConfigChange() {

}

// OnShutDown 服务被终止的通知
func OnShutDown() {

}
