package trace

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryTraceServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (result any, err error) {
	names := strings.Split(strings.Trim(info.FullMethod, "/"), "/")
	serviceName := names[0]
	methodName := names[1]
	provider, ok := providers[serviceName]
	if !ok {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	carrier := MetadataTextMapCarrier(md)

	err = ServerEndTrace(ctx, serviceName, methodName, carrier, provider, func(ctx context.Context) (err error) {
		result, err = handler(ctx, req)
		return
	})
	return
}

func StreamTraceServerInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	if info.FullMethod == "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo" {
		return handler(srv, ss)
	}
	names := strings.Split(strings.Trim(info.FullMethod, "/"), "/")
	serviceName := names[0]
	methodName := names[1]
	provider, ok := providers[serviceName]
	if !ok {
		return handler(srv, ss)
	}
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		md = metadata.MD{}
	}

	carrier := MetadataTextMapCarrier(md)

	return ServerEndTrace(ss.Context(), serviceName, methodName, carrier, provider, func(ctx context.Context) error {
		return handler(srv, &wrapper{ServerStream: ss, ctx: ctx})
	})
}

func UnaryTraceClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	var srcNames []string
	if srcMethod, ok := grpc.Method(ctx); ok {
		srcNames = strings.Split(strings.Trim(srcMethod, "/"), "/")
	} else { //非grpc没有svr方法,比如grpc-gateway
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	srcServiceName := srcNames[0]
	provider, ok := providers[srcServiceName]
	if !ok {
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	carrier := MetadataTextMapCarrier(md)

	return ClientEndTrace(ctx, method, carrier, provider, func(ctx context.Context) error {
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}

func StreamTraceClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
	var srcNames []string
	if srcMethod, ok := grpc.Method(ctx); ok {
		srcNames = strings.Split(strings.Trim(srcMethod, "/"), "/")
	} else { //非grpc没有svr方法,比如grpc-gateway
		return streamer(ctx, desc, cc, method, opts...)
	}
	srcServiceName := srcNames[0]
	provider, ok := providers[srcServiceName]
	if !ok {
		return streamer(ctx, desc, cc, method, opts...)
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	carrier := MetadataTextMapCarrier(md)

	err = ClientEndTrace(ctx, method, carrier, provider, func(ctx context.Context) (err error) {
		ctx = metadata.NewOutgoingContext(ctx, md)
		cs, err = streamer(ctx, desc, cc, method, opts...)
		return
	})
	return
}

type MetadataTextMapCarrier metadata.MD

// Get returns the value associated with the passed key.
func (m MetadataTextMapCarrier) Get(key string) string {
	g, ok := m[key]
	if !ok || len(g) == 0 {
		return ""
	}
	return g[0]
}

// Set stores the key-value pair.
func (m MetadataTextMapCarrier) Set(key string, value string) {
	m[key] = []string{value}
}

// Keys lists the keys stored in this carrier.
func (m MetadataTextMapCarrier) Keys() []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

type wrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrapper) Context() context.Context {
	return w.ctx
}
