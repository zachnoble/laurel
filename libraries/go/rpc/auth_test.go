package rpc

import (
	"context"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"

	authv1 "github.com/zachnoble/laurel/rpc/generated/go/auth/v1"
	"github.com/zachnoble/laurel/rpc/generated/go/auth/v1/authv1connect"
)

// The real AuthInternalService/Authenticate descriptor: internal, public
// (no user needed).
func authenticateSpec() connect.Spec {
	return connect.Spec{
		Procedure: authv1connect.AuthInternalServiceAuthenticateProcedure,
		Schema:    authv1.File_auth_v1_auth_proto.Services().ByName("AuthInternalService").Methods().ByName("Authenticate"),
	}
}

func TestInternalProcedureRejectsGatewayTraffic(t *testing.T) {
	_, users := newKeys(t, time.Minute)
	a := NewAuthenticator(users)
	spec := authenticateSpec()

	if _, err := a.authenticate(context.Background(), spec, http.Header{ViaGatewayHeader: {"1"}}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("through the gateway: %v, want PermissionDenied", err)
	}
	if _, err := a.authenticate(context.Background(), spec, http.Header{}); err != nil {
		t.Errorf("from another service: %v, want allowed", err)
	}
}

func TestParsesPEMKeys(t *testing.T) {
	const privPEM = `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIAGzy9f3Q9Y5Oku91MzYwFHuAXwixDisioGuavsSdAt2
-----END PRIVATE KEY-----`
	const pubPEM = `-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEA6HV2QGObTU2/I+l39kOu5tPacbrHZhWjXyMd2n6y/Dk=
-----END PUBLIC KEY-----`
	priv, err := ParsePrivateKey(privPEM)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := ParsePublicKey(pubPEM)
	if err != nil {
		t.Fatal(err)
	}
	if !pub.Equal(priv.Public()) {
		t.Error("PEM key pair does not match")
	}
}
