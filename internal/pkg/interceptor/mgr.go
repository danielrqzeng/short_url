// gen by iyfiysi at 2021 May 19

package interceptor

import (
	"github.com/spf13/viper"
	grpcPB "google.golang.org/grpc"
	grpcInterceptor "iyfiysi.com/short_url/internal/pkg/interceptor/grpc"
	httpInterceptor "iyfiysi.com/short_url/internal/pkg/interceptor/http"
	"net/http"
	"sync"
)

var (
	MgrInstance *MgrInstanceType
	MgrOnce     sync.Once
)

//Mgr 拦截器管理实例
func Mgr() *MgrInstanceType {
	MgrOnce.Do(func() {
		MgrInstance = &MgrInstanceType{}
		MgrInstance.Init()
	})
	return MgrInstance
}

//Mgr 拦截器
type MgrInstanceType struct {
}

//Init 初始化
func (mgr *MgrInstanceType) Init() {
}

// GetServerInterceptors all interceptors for server
func (mgr *MgrInstanceType) GetServerInterceptors() (
	interceptors grpcPB.UnaryServerInterceptor) {
	return grpcInterceptor.InterceptorMgr().GetServerInterceptors()
}

// GetGatewayInterceptors all interceptors for gateway
func (mgr *MgrInstanceType) GetGatewayInterceptors() (
	interceptors grpcPB.UnaryClientInterceptor) {
	return grpcInterceptor.InterceptorMgr().GetGatewayInterceptors()
}

// GetHttpInterceptors all interceptors for http
func (mgr *MgrInstanceType) GetHttpInterceptors(h http.Handler,
) func(w http.ResponseWriter, r *http.Request) {
	if viper.GetBool("metrics.enable") {
		httpInterceptor.InterceptorMgr().Use(httpInterceptor.Metrics)
	}
	httpInterceptor.InterceptorMgr().Use(httpInterceptor.Cors)
	httpInterceptor.InterceptorMgr().Use(httpInterceptor.Trace)
	httpInterceptor.InterceptorMgr().Use(httpInterceptor.Query)
	return httpInterceptor.InterceptorMgr().Handler(h)
}
