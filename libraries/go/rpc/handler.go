package rpc

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/validate"

	"github.com/zachnoble/laurel/libraries/go/logging"
)

// HandlerOptions returns the standard options for mounting a Connect handler:
// public error mapping, then the given interceptors (e.g. auth), then
// protovalidate, plus panic recovery.
func HandlerOptions(interceptors ...connect.Interceptor) []connect.HandlerOption {
	chain := append([]connect.Interceptor{Errors{}}, interceptors...)
	chain = append(chain, validate.NewInterceptor())

	return []connect.HandlerOption{
		connect.WithInterceptors(chain...),
		connect.WithRecover(func(ctx context.Context, spec connect.Spec, _ http.Header, p any) error {
			logging.FromContext(ctx).Error("panic recovered", "procedure", spec.Procedure, "error", p)
			return PublicError(ctx, fmt.Errorf("panic: %v", p))
		}),
	}
}
