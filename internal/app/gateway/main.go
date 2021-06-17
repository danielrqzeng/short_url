// gen by iyfiysi at 2021 May 19

package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	etcdNaming "github.com/coreos/etcd/clientv3/naming"
	"github.com/golang/protobuf/proto"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"iyfiysi.com/short_url/internal/app/gateway/service"
	"iyfiysi.com/short_url/internal/pkg/data"
	"iyfiysi.com/short_url/internal/pkg/governance"
	"iyfiysi.com/short_url/internal/pkg/interceptor"
	grpcInterceptor "iyfiysi.com/short_url/internal/pkg/interceptor/grpc"
	"iyfiysi.com/short_url/internal/pkg/logger"
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
	swaggerAddr string // 文档侦听地址，格式如：127.0.0.1:8080
}

//Init ...
func (app *ApplicationType) Init() {
}

func writeRoot(w http.ResponseWriter) (err error) {
	rootIndexFile := viper.GetString("indexFile")
	//byteFile, err := utils.ReadFileAsByte(rootIndexFile)
	//if err != nil {
	//	logger.MainLogger.Error(err.Error())
	//	return err
	//}
	//w.Write(byteFile)

	statikFS, err := fs.New()
	if err != nil {
		logger.MainLogger.Error(err.Error())
		return
	}
	r, err := statikFS.Open("/" + rootIndexFile)
	if err != nil {
		logger.MainLogger.Error(err.Error())
		return
	}
	defer r.Close()
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		logger.MainLogger.Error(err.Error())
		return
	}

	w.Write(contents)
	return nil
}

//responseHeaderMatcher 相应头部,将decode出来的，做成重定向返回
func responseHeaderMatcher(
	ctx context.Context, w http.ResponseWriter, rsp proto.Message) error {
	logger.MainLogger.Error("responseHeaderMatcher mark")
	headers := w.Header()
	if location, ok := headers["Grpc-Metadata-Location"]; ok {
		w.Header().Set("Location", location[0])
		w.WriteHeader(http.StatusFound)
	}

	return nil
}

//OnProtoErrorHandlerFunc pb方法报错时候，进入此处处理
func OnProtoErrorHandlerFunc(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	request *http.Request, e error) {

	//"rpc error: code = Unknown desc = not valid link"
	grpcErr := status.Convert(e)
	//定向到主页
	if grpcErr.Message() == data.IndexRequestErr {
		w.WriteHeader(http.StatusOK)
		w.Header().Del("Content-Type")
		w.Header().Add("Content-Type", "text/html;charset=utf-8")
		writeRoot(w)
		return
	}
	//其他的错误，返回json
	rsp := &data.BaseResponse{}
	rsp.RetCode = -1
	rsp.RetMsg = e.Error()
	rsp.MsgShow = grpcErr.Message()
	byteStr, err := json.Marshal(rsp)

	if err == nil {
		w.Write(byteStr)
	}
	return
}

//grpcServer ...
func (app *ApplicationType) grpcServer() (gwMux *runtime.ServeMux) {
	gwMux = runtime.NewServeMux(
		runtime.WithForwardResponseOption(responseHeaderMatcher),
		runtime.WithProtoErrorHandler(OnProtoErrorHandlerFunc))

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

	//注册-监控主键
	e := &governance.EtcdType{}
	cli, err := governance.DefaultEtcdV3Client()
	if err != nil {
		panic(err)
	}
	metricKey := viper.GetString("etcd.metricKey") + "/gateway/" + instance
	err = e.RunToKeepAlive(cli, metricKey,
		func(leaseID clientv3.LeaseID) {
			ops := clientv3.WithLease(leaseID)
			_, err := cli.Put(context.TODO(), metricKey, instance, ops)
			if err != nil {
				panic(fmt.Sprintf("RunToKeepAlive key=%s err,err=%s", metricKey, err.Error()))
			}
		},
		func(code int, msg string) {
			panic(fmt.Sprintf("RunToKeepAlive key=%s err,code=%d,msg=%s", metricKey, code, msg))
		})

	if err != nil {
		panic(fmt.Sprintf("RunToKeepAlive key=%s err,err=%s", metricKey, err.Error()))
		return
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

//runSwaggerHTTP swagger服务
func (app *ApplicationType) runSwaggerHTTP() {
	if !viper.GetBool("swagger.enable") {
		return
	}

	instance, err := governance.GetSetupInstanceAddrByConfKey("swagger")
	if err != nil {
		panic(err)
	}
	app.swaggerAddr = instance

	//注册-监控主键
	e := &governance.EtcdType{}
	cli, err := governance.DefaultEtcdV3Client()
	if err != nil {
		panic(err)
	}
	etcdKey := viper.GetString("etcd.swaggerKey") + instance
	err = e.RunToKeepAlive(cli, etcdKey,
		func(leaseID clientv3.LeaseID) {
			ops := clientv3.WithLease(leaseID)
			_, err := cli.Put(context.TODO(), etcdKey, instance, ops)
			if err != nil {
				panic(fmt.Sprintf("RunToKeepAlive key=%s err,err=%s", etcdKey, err.Error()))
			}
		},
		func(code int, msg string) {
			panic(fmt.Sprintf("RunToKeepAlive key=%s err,code=%d,msg=%s", etcdKey, code, msg))
		})

	if err != nil {
		panic(fmt.Sprintf("RunToKeepAlive key=%s err,err=%s", etcdKey, err.Error()))
		return
	}

	//swagger
	statikFS, err := fs.New()
	if err != nil {
		return
	}
	app.grpcServer()
	mux := http.NewServeMux()
	mux.Handle("/", app.grpcServer())
	swaggerPath := viper.GetString("swagger.path")
	mux.Handle(swaggerPath,
		http.StripPrefix(swaggerPath, http.FileServer(statikFS)))
	go http.ListenAndServe(instance, mux)
	return
}

func (app *ApplicationType) Run() (err error) {
	//metrics
	go app.runMetricsHTTP()
	go app.runSwaggerHTTP()

	//grpc
	err = app.runGRPC()
	if err != nil {
		return
	}
	return
}
