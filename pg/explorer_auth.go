package pg

import (
	"context"

	"google.golang.org/grpc"
)

func (e *Explorer) UnaryAuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	// TODO

	return handler(ctx, req)
}
