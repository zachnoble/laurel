package rpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/zachnoble/laurel/libraries/go/logging"
)

var errInternal = errors.New("An unexpected error occurred")

// PublicError passes *connect.Error through unchanged and logs anything else,
// replacing it with a generic internal error so details never reach clients.
func PublicError(ctx context.Context, err error) error {
	if _, ok := errors.AsType[*connect.Error](err); ok {
		return err
	}
	switch {
	case errors.Is(err, context.Canceled):
		return connect.NewError(connect.CodeCanceled, err)
	case errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeDeadlineExceeded, err)
	}

	logging.FromContext(ctx).Error("internal error", "error", err)
	return connect.NewError(connect.CodeInternal, errInternal)
}

// Errors is an interceptor that applies PublicError to handler errors.
type Errors struct{}

func (Errors) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		res, err := next(ctx, req)
		if err != nil {
			return nil, PublicError(ctx, err)
		}
		return res, nil
	}
}

func (Errors) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (Errors) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if err := next(ctx, conn); err != nil {
			return PublicError(ctx, err)
		}
		return nil
	}
}
