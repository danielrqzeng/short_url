// gen by iyfiysi at 2021 May 19

package gateway

import (
	"context"
	etcdNaming "github.com/coreos/etcd/clientv3/naming"
	"github.com/golang/protobuf/proto"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"iyfiysi.com/short_url/internal/app/gateway/service"
	"iyfiysi.com/short_url/internal/pkg/governance"
	"iyfiysi.com/short_url/internal/pkg/interceptor"
	grpcInterceptor "iyfiysi.com/short_url/internal/pkg/interceptor/grpc"
	"iyfiysi.com/short_url/internal/pkg/trace"
	"iyfiysi.com/short_url/internal/pkg/utils"
	"net/http"
	"sync"
)

var (
	appSingleton *ApplicationType
	appOnce      sync.Once
)

//Mgr 拦截器管理实例
func App() *ApplicationType {
	appOnce.Do(func() {
		appSingleton = &ApplicationType{}
		appSingleton.Init()
	})
	return appSingleton
}

//ApplicationType gateway app定义
type ApplicationType struct {
	serviceAddr string // 侦听地址，格式如：127.0.0.1:8000
	metricAddr  string // 监控侦听地址，格式如：127.0.0.1:8000
}

//Init ...
func (app *ApplicationType) Init() {
}

// responseHeaderMatcher 302重定向
func responseHeaderMatcher(
	ctx context.Context, w http.ResponseWriter, rsp proto.Message) error {
	headers := w.Header()
	if location, ok := headers["Grpc-Metadata-Location"]; ok {
		w.Header().Set("Location", location[0])
		w.WriteHeader(http.StatusFound)
	}

	return nil
}

//grpcServer ...
func (app *ApplicationType) grpcServer() (gwMux *runtime.ServeMux) {
	gwMux = runtime.NewServeMux(runtime.WithForwardResponseOption(responseHeaderMatcher))

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel
	//defer cancel()

	//----------etcd的服务发现option-------------------
	cli, err := governance.DefaultEtcdV3Client()
	if err != nil {
		panic(err)
	}
	//defer cli.Close()

	//--------------服务发现&负债均衡组件----------------------
	r := &etcdNaming.GRPCResolver{Client: cli} //其需要配合gw.RegisterXXXXXServiceHandlerFromEndpoint中的endpoint参数使用
	lbOption := grpc.WithBalancer(grpc.RoundRobin(r))

	//--------------ssl证书option----------------------
	serverName := viper.GetString("keystore.serverName")
	caFile := viper.GetString("keystore.ca")
	privateFile := viper.GetString("keystore.private")
	publicFile := viper.GetString("keystore.public")
	_, clientCred, err := utils.GenCredentials(caFile, publicFile, privateFile, serverName)
	if err != nil {
		panic(err)
	}
	sslOption := grpc.WithTransportCredentials(clientCred)
	//--------------拦截器之服务调用鉴权----------------------
	tokenOption := grpc.WithPerRPCCredentials(grpcInterceptor.BearerRPCCredentials()) //调用认证

	//--------------拦截器option----------------------
	interceptors := interceptor.Mgr().GetGatewayInterceptors()
	interceptorOption := grpc.WithUnaryInterceptor(
		grpcMiddleware.ChainUnaryClient(interceptors))

	//所有选项
	opts := []grpc.DialOption{
		lbOption,
		sslOption,
		tokenOption,
		interceptorOption,
	}

	serviceKey := viper.GetString("etcd.serviceKey")
	err = service.DoRegister(ctx, serviceKey, gwMux, opts)
	if err != nil {
		return
	}
	return
}

//runGRPC grpc服务
func (app *ApplicationType) runGRPC() (err error) {
	instance, err := governance.GetSetupInstanceAddrByConfKey("gateway")
	if err != nil {
		return
	}
	// 将gateway的服务侦听地址设置到viper中（以备其他地方使用），key为listen
	app.serviceAddr = instance
	viper.Set("listen", app.serviceAddr)
	trace.Init() // 对opentracing.GlobalTracer() 重新初始化，使得侦听实例在trace的tag中生效

	gwMux := app.grpcServer()
	HTTPMux := http.NewServeMux()
	HTTPMux.HandleFunc("/", interceptor.Mgr().GetHttpInterceptors(gwMux))

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	err = http.ListenAndServe(instance, HTTPMux)
	return
}

//runMetricsHTTP metric服务
func (app *ApplicationType) runMetricsHTTP() {
	if !viper.GetBool("metrics.enable") {
		return
	}

	instance, err := governance.GetSetupInstanceAddrByConfKey("metrics.gateway")
	if err != nil {
		panic(err)
	}
	app.metricAddr = instance
	metricsPath := viper.GetString("metrics.gateway.path")
	HTTPMux := http.NewServeMux()
	HTTPMux.Handle(metricsPath, promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{},
	))

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	err = http.ListenAndServe(instance, HTTPMux)
	return
}

func (app *ApplicationType) Run() (err error) {
	//metrics
	go app.runMetricsHTTP()

	//grpc
	err = app.runGRPC()
	if err != nil {
		return
	}
	return
}
