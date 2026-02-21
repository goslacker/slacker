package interceptor

import (
	"buf.build/go/protovalidate"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type streamWrapper struct {
	grpc.ServerStream
}

func (w *streamWrapper) RecvMsg(m any) (err error) {
	err = w.ServerStream.RecvMsg(m)
	if err != nil {
		return
	}
	err = validate(m)
	return
}

func StreamValidateInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	return handler(srv, &streamWrapper{ss})
}

func UnaryValidateInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (result any, err error) {
	err = validate(req)
	if err != nil {
		return
	}

	// 调用被拦截的方法
	return handler(ctx, req)
}

func validate(req any) (err error) {
	if req != nil {
		if err = protovalidate.Validate(req.(proto.Message)); err != nil {
			err = status.New(codes.InvalidArgument, err.Error()).Err()
			return
		}
	}
	return
}
