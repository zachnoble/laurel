package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"

	"github.com/zachnoble/laurel/libraries/go/logging"
	"github.com/zachnoble/laurel/libraries/go/rpc"
)

// procedurePath matches Connect procedure paths, /<name>.v<N>.<Service>/<Method>.
// <name> is the upstream service, so new services need no gateway config.
// Service names are lowercase letters and digits (enforced by tools/repo).
var procedurePath = regexp.MustCompile(`^/([a-z][a-z0-9]*)\.v[0-9]+\.[A-Za-z][A-Za-z0-9]*/[A-Za-z][A-Za-z0-9]*$`)

// Gateway proxies browser RPCs to services, attaching the caller's identity.
//
// It does not decide which procedures are reachable: services enforce the
// access rules declared in their protos (rpc.Authenticator). Every request
// it forwards carries rpc.ViaGatewayHeader, which internal services reject,
// so they can't be called through it.
type Gateway struct {
	upstream       func(name string) string
	sessionService string
	sessions       *Sessions
	proxy          *httputil.ReverseProxy
}

// New takes upstream, which maps a service name to its base URL.
func New(upstream func(name string) string, transport http.RoundTripper, sessionService string, sessions *Sessions) *Gateway {
	g := &Gateway{upstream: upstream, sessionService: sessionService, sessions: sessions}
	g.proxy = &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			target, _ := url.Parse(g.upstream(upstreamName(pr.In.URL.Path)))
			pr.SetURL(target)
			pr.SetXForwarded()
		},
		Transport:     transport,
		FlushInterval: -1, // stream responses through immediately
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			name := upstreamName(r.URL.Path)
			logging.FromContext(r.Context()).Error("upstream error", "service", name, "path", r.URL.Path, "error", err)
			// A Connect error, so clients show the message. Locally this usually
			// means the service wasn't started: `just up <name>`.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"code":    "unavailable",
				"message": fmt.Sprintf("Service %q is unavailable (not running locally? run `just up %s`)", name, name),
			})
		},
	}
	return g
}

func upstreamName(path string) string {
	if m := procedurePath.FindStringSubmatch(path); m != nil {
		return m[1]
	}
	return ""
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := upstreamName(r.URL.Path)
	if name == "" {
		http.NotFound(w, r)
		return
	}

	// Only the gateway may assert identity (the service middleware already
	// drops these at the edge; this keeps the handler safe on its own).
	r.Header.Del(rpc.IdentityHeader)
	r.Header.Set(rpc.ViaGatewayHeader, "1")

	if cookie, err := r.Cookie(g.sessions.cookieName); err == nil && cookie.Value != "" {
		token, err := g.sessions.Token(r.Context(), cookie.Value)
		if err != nil {
			logging.FromContext(r.Context()).Error("resolve session", "error", err)
			http.Error(w, "session service unavailable", http.StatusServiceUnavailable)
			return
		}
		if token != "" {
			r.Header.Set(rpc.IdentityHeader, token)
		}
	}
	// Only the session service sees cookies.
	if name != g.sessionService {
		r.Header.Del("Cookie")
	}

	g.proxy.ServeHTTP(w, r)
}
