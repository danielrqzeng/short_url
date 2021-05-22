// Code generated by protoc-gen-iyfiysi at 2021 May 22
// DO NOT EDIT.

package service

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	gw "iyfiysi.com/short_url/proto"
)

func DoRegister(
	ctx context.Context,
	serviceKey string,
	mux *runtime.ServeMux,
	opts []grpc.DialOption) (err error) {
	err = gw.RegisterShortUrlServiceHandlerFromEndpoint(ctx, mux, serviceKey, opts)
	if err != nil {
		return
	}

	return
}
