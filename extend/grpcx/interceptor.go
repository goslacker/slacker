package grpcx

import (
	"context"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func validateInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (result any, err error) {
	if req != nil {
		var v *protovalidate.Validator
		v, err = protovalidate.New()
		if err != nil {
			err = status.New(codes.Internal, "new validator failed").Err()
			return
		}
		if err = v.Validate(req.(proto.Message)); err != nil {
			err = status.New(codes.InvalidArgument, err.Error()).Err()
			return
		}
	}

	// 调用被拦截的方法
	return handler(ctx, req)
}
