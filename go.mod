module iyfiysi.com/short_url

go 1.13

require (
	github.com/RQZeng/num-shuffle v0.0.0-20210506033248-28794bf465bd
	github.com/adamzy/cedar-go v0.0.0-20170805034717-80a9c64b256d
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/bwmarrin/snowflake v0.3.0
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-redsync/redsync v1.4.2
	github.com/go-sql-driver/mysql v1.6.0
	//github.com/gohouse/gorose v1.0.5
	github.com/gohouse/gorose/v2 v2.1.4-rc.0.20191228024627-ac07c107a0cd
	github.com/golang/protobuf v1.4.3
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.1.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.13.0
	github.com/json-iterator/go v1.1.11
	github.com/opentracing/opentracing-go v1.2.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/platinummonkey/go-concurrency-limits v0.5.5
	github.com/prometheus/client_golang v0.9.3
	github.com/robfig/cron/v3 v3.0.1
	github.com/spf13/viper v1.7.1
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	go.uber.org/zap v1.15.0
	golang.org/x/net v0.0.0-20191002035440-2ec189313ef0
	google.golang.org/genproto v0.0.0-20191108220845-16a3f7862a1a
	google.golang.org/grpc v1.36.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace iyfiysi.com/short_url => ./

replace github.com/golang/protobuf v1.4.3 => github.com/golang/protobuf v1.3.2

replace google.golang.org/grpc v1.36.0 => google.golang.org/grpc v1.26.0

replace github.com/gohouse/e v0.0.3-rc.0.20200724104652-25ebf8c9c305 => github.com/gohouse/e v0.0.0-20200724104652-25ebf8c9c305
