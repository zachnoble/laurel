package gateway

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/sync/singleflight"

	authv1 "github.com/zachnoble/laurel/rpc/generated/go/auth/v1"
	"github.com/zachnoble/laurel/rpc/generated/go/auth/v1/authv1connect"
)

// Sessions exchanges session cookies for identity tokens via the auth
// service, caching results for ttl. A revoked session therefore keeps working
// for at most ttl; identity tokens themselves live longer (see auth).
//
// The cache is keyed by the session cookie's value alone (other cookies don't
// fragment it), holds at most size entries (least recently used are evicted,
// so random cookies can't grow memory), and makes one auth call per cookie
// at a time however many requests are waiting on it.
type Sessions struct {
	client     authv1connect.AuthInternalServiceClient
	cookieName string

	cache *expirable.LRU[string, string] // session cookie value -> token ("" = no session)
	group singleflight.Group
}

func NewSessions(client authv1connect.AuthInternalServiceClient, cookieName string, size int, ttl time.Duration) *Sessions {
	return &Sessions{
		client:     client,
		cookieName: cookieName,
		cache:      expirable.NewLRU[string, string](size, nil, ttl),
	}
}

// Token returns the identity token for a session cookie value, or "" if it
// is not a valid session. An error means auth could not be reached.
func (s *Sessions) Token(ctx context.Context, session string) (string, error) {
	if token, ok := s.cache.Get(session); ok {
		return token, nil
	}

	v, err, _ := s.group.Do(session, func() (any, error) {
		// Not tied to the first caller: others may be waiting on the result.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		var token string
		res, err := s.client.Authenticate(ctx, connect.NewRequest(&authv1.AuthenticateRequest{Cookie: s.cookieName + "=" + session}))
		switch {
		case err == nil:
			token = res.Msg.GetToken()
		case connect.CodeOf(err) == connect.CodeUnauthenticated:
			// No session: cache the miss too, so bad cookies don't hit auth every request.
		default:
			return "", err
		}
		s.cache.Add(session, token)
		return token, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}
