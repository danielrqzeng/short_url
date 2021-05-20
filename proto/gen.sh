#!/bin/sh

protoc -I. --go_out=plugins=grpc:. ./service.proto
protoc -I. --grpc-gateway_out=logtostderr=true:. ./service.proto
protoc -I. --iyfiysi_out=domain=iyfiysi.com,app=short_url:. ./service.proto