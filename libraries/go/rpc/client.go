package rpc

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// HTTPClient returns a client for calling other services in the cluster over
// unencrypted HTTP/2 (h2c), which every server from libraries/go/server
// accepts. It propagates trace context (W3C traceparent) and records client
// spans. It has no overall timeout, so streams work; unary calls get a
// deadline from Clients' default timeout instead.
func HTTPClient() *http.Client {
	return &http.Client{Transport: Transport()}
}

// Transport is HTTPClient's round tripper, for proxies.
func Transport() http.RoundTripper {
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	return otelhttp.NewTransport(&http.Transport{
		Protocols:           protocols,
		DialContext:         (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		MaxIdleConnsPerHost: 64,
		IdleConnTimeout:     90 * time.Second,
	}, otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string { return r.URL.Path }))
}

// ServiceURL resolves a service's base URL from a template such as
// "http://{name}:8080" (ECS Service Connect and docker compose both expose
// services by name).
func ServiceURL(template, name string) string {
	return strings.ReplaceAll(template, "{name}", name)
}

// Clients builds Connect clients for calling other services with the
// standard behaviour: the caller's identity forwarded and a default deadline.
//
//	orders := rpc.NewClient(clients, "orders", ordersv1connect.NewOrdersServiceClient)
type Clients struct {
	HTTP *http.Client

	urlTemplate string
	overrides   map[string]string
	timeout     time.Duration
}

type ClientsConfig struct {
	// URLTemplate, e.g. "http://{name}:8080".
	URLTemplate string
	// Overrides maps a service name to a base URL, replacing the template
	// (e.g. to point one dependency at a shared dev environment).
	Overrides map[string]string
	// Timeout for calls whose context has no deadline. Default 10s.
	Timeout time.Duration
}

func NewClients(cfg ClientsConfig) *Clients {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &Clients{
		HTTP:        HTTPClient(),
		urlTemplate: cfg.URLTemplate,
		overrides:   cfg.Overrides,
		timeout:     cfg.Timeout,
	}
}

// URL returns the base URL of the named service.
func (c *Clients) URL(name string) string {
	if u, ok := c.overrides[name]; ok {
		return u
	}
	return ServiceURL(c.urlTemplate, name)
}

// Options returns the client options for calling the named service.
func (c *Clients) Options(name string) []connect.ClientOption {
	return []connect.ClientOption{
		connect.WithInterceptors(
			defaultTimeout{c.timeout},
			ForwardIdentity{},
		),
	}
}

// NewClient builds a client for the named service with a generated
// constructor, e.g. authv1connect.NewAuthInternalServiceClient.
func NewClient[T any](c *Clients, name string, newClient func(connect.HTTPClient, string, ...connect.ClientOption) T) T {
	return newClient(c.HTTP, c.URL(name), c.Options(name)...)
}

// defaultTimeout gives unary calls a deadline if the caller set none. The
// deadline propagates downstream (Connect-Timeout-Ms), so a whole call tree
// shares the original budget.
type defaultTimeout struct{ d time.Duration }

func (t defaultTimeout) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if _, ok := ctx.Deadline(); !ok && req.Spec().IsClient {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, t.d)
			defer cancel()
		}
		return next(ctx, req)
	}
}

func (defaultTimeout) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (defaultTimeout) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}
