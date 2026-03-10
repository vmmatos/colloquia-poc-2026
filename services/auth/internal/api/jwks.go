package api

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"net/http"

	"github.com/gin-gonic/gin"
)

// jwksHandler returns a Gin handler that serves the RSA public key as a JSON
// Web Key Set (JWKS). This endpoint is consumed by KrakenD (and any other
// service that needs to verify RS256 JWTs issued by the auth service) at
// startup and periodically to refresh the cached key material.
//
// Path: GET /.well-known/jwks.json
func jwksHandler(pub *rsa.PublicKey) gin.HandlerFunc {
	// Pre-compute the JWK once at startup — the key never changes at runtime.
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())

	eRaw := make([]byte, 4)
	binary.BigEndian.PutUint32(eRaw, uint32(pub.E))
	// Trim leading zero bytes (big-endian, RFC 7518 §2).
	start := 0
	for start < len(eRaw)-1 && eRaw[start] == 0 {
		start++
	}
	e := base64.RawURLEncoding.EncodeToString(eRaw[start:])

	body := gin.H{
		"keys": []gin.H{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": "colloquia-auth-key-1",
				"n":   n,
				"e":   e,
			},
		},
	}

	return func(c *gin.Context) {
		// Allow aggressive caching by API-gateway / CDN — key rotations are
		// intentional and deployed explicitly.
		c.Header("Cache-Control", "public, max-age=3600")
		c.JSON(http.StatusOK, body)
	}
}
