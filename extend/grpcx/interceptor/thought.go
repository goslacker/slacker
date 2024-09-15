package interceptor

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryThoughtClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	return invoker(thought(ctx), method, req, reply, cc, opts...)
}

func StreamThoughtClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
	return streamer(thought(ctx), desc, cc, method, opts...)
}

func thought(ctx context.Context) (newCtx context.Context) {
	md, _ := metadata.FromIncomingContext(ctx)
	outmd, _ := metadata.FromOutgoingContext(ctx)

	thoughtKeys := []string{"authorization"}

	for _, key := range thoughtKeys {
		if len(outmd[key]) == 0 {
			outmd[key] = md[key]
		}
	}

	newCtx = metadata.NewOutgoingContext(ctx, outmd)

	return
}
