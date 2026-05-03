package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func parseJWK(secret string) (interface{}, error) {
	var jwk struct {
		X   string `json:"x"`
		Y   string `json:"y"`
		Kty string `json:"kty"`
		Crv string `json:"crv"`
	}

	if err := json.Unmarshal([]byte(secret), &jwk); err != nil {
		return nil, err
	}

	if jwk.Kty != "EC" || jwk.Crv != "P-256" {
		return nil, fmt.Errorf("unsupported JWK: kty=%s, crv=%s", jwk.Kty, jwk.Crv)
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode x: %v", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode y: %v", err)
	}

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			alg, _ := token.Header["alg"].(string)
			
			// 1. If it's an asymmetric algorithm (ES256, RS256), we MUST return a Public Key
			if strings.HasPrefix(alg, "ES") || strings.HasPrefix(alg, "RS") {
				// Try JWK (JSON)
				if key, err := parseJWK(jwtSecret); err == nil {
					return key, nil
				}

				// Try PEM
				formattedKey := strings.ReplaceAll(jwtSecret, "\\n", "\n")
				if key, err := jwt.ParseECPublicKeyFromPEM([]byte(formattedKey)); err == nil {
					return key, nil
				}

				return nil, fmt.Errorf("could not parse public key for %s algorithm. Check SUPABASE_JWT_SECRET format", alg)
			}

			// 2. Fallback to HS256 (Symmetric)
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			log.Printf("JWT Validation Failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
