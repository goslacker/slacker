package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryThroughClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	return invoker(through(ctx), method, req, reply, cc, opts...)
}

func StreamThroughClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
	return streamer(through(ctx), desc, cc, method, opts...)
}

var throughKeys = []string{}

func SetThroughKeys(keys ...string) {
	throughKeys = append(throughKeys, keys...)
}

func through(ctx context.Context) (newCtx context.Context) {
	md, _ := metadata.FromIncomingContext(ctx)
	outmd, _ := metadata.FromOutgoingContext(ctx)
	if outmd == nil {
		outmd = make(metadata.MD)
	}

	for _, key := range throughKeys {
		if len(outmd[key]) == 0 {
			outmd[key] = md[key]
		}
	}

	newCtx = metadata.NewOutgoingContext(ctx, outmd)

	return
}
