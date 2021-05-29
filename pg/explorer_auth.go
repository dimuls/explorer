package pg

import (
	"context"
	"explorer"

	"google.golang.org/grpc"
)

func (e *Explorer) PostLogin(ctx context.Context, in *explorer.PostLoginReq, opts ...grpc.CallOption) (*explorer.PostLoginRes, error) {
	return nil, nil
}
