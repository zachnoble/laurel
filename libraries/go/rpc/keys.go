package rpc

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"
)

// Keys are Ed25519, in either form:
//   - base64 of the raw 32-byte seed / public key (`just keygen`), or
//   - PEM: PKCS#8 private keys and PKIX public keys (Terraform's tls_private_key).

var (
	errPrivateKey = errors.New("private key must be a base64 32-byte ed25519 seed or a PKCS#8 PEM ed25519 key")
	errPublicKey  = errors.New("public key must be a base64 32-byte ed25519 key or a PKIX PEM ed25519 key")
)

func ParsePrivateKey(s string) (ed25519.PrivateKey, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "-----BEGIN") {
		block, _ := pem.Decode([]byte(s))
		if block == nil {
			return nil, errPrivateKey
		}
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, errPrivateKey
		}
		if k, ok := key.(ed25519.PrivateKey); ok {
			return k, nil
		}
		return nil, errPrivateKey
	}
	seed, err := base64.StdEncoding.DecodeString(s)
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, errPrivateKey
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

func ParsePublicKey(s string) (ed25519.PublicKey, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "-----BEGIN") {
		block, _ := pem.Decode([]byte(s))
		if block == nil {
			return nil, errPublicKey
		}
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, errPublicKey
		}
		if k, ok := key.(ed25519.PublicKey); ok {
			return k, nil
		}
		return nil, errPublicKey
	}
	key, err := base64.StdEncoding.DecodeString(s)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return nil, errPublicKey
	}
	return ed25519.PublicKey(key), nil
}
