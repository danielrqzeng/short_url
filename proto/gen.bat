@echo on
protoc.exe -I. --go_out=plugins=grpc:. ./service.proto
protoc.exe -I. --grpc-gateway_out=logtostderr=true:. ./service.proto
protoc.exe -I. --swagger_out=logtostderr=true:../swagger ./service.proto
protoc.exe -I. --iyfiysi_out=domain=iyfiysi.com,app=short_url:. ./service.proto