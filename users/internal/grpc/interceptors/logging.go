package interceptors

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingUnaryInterceptor logs gRPC requests and responses with timing information
func LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Log incoming request
		log.Printf("gRPC Request: method=%s", info.FullMethod)

		// Call the handler
		resp, err := handler(ctx, req)

		// Log response with timing and status
		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			if grpcErr, ok := status.FromError(err); ok {
				code = grpcErr.Code()
			} else {
				code = codes.Internal
			}
		}

		log.Printf("gRPC Response: method=%s, code=%s, duration=%v",
			info.FullMethod, code.String(), duration)

		// Log errors for non-OK status codes (except NotFound which might be expected)
		if err != nil && code != codes.NotFound {
			log.Printf("gRPC Error: method=%s, error=%v", info.FullMethod, err)
		}

		return resp, err
	}
}

// RecoveryUnaryInterceptor catches panics and converts them to gRPC errors
func RecoveryUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("gRPC Panic recovered: method=%s, panic=%v", info.FullMethod, r)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}