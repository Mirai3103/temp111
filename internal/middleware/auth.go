// Package middleware provides reusable HTTP middleware for the API gateway.
package middleware

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"

	"github.com/FPT-OJT/minstant-ai.git/internal/constants"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth returns a middleware that verifies RS256 JWTs in the Authorization
// header using the provided RSA public key.
//
// Behavior:
//   - If NO Authorization header is present, it allows the request through
//     unauthenticated (upstream services handle public routes).
//   - If a Bearer token IS present, it must be valid and unexpired.
//   - If invalid, returns 401 Unauthorized immediately.
//   - If valid, extracts the "sub" claim, injects it into the request context,
//     and adds the "X-User-Id" HTTP header for upstream services.
func JWTAuth(pubKey *rsa.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				sendUnauthorized(w, "Malformed Authorization header. Expected 'Bearer <token>'")
				return
			}
			tokenStr := parts[1]

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return pubKey, nil
			})

			if err != nil || !token.Valid {
				sendUnauthorized(w, "Invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				sendUnauthorized(w, "Invalid token claims")
				return
			}

			sub, err := claims.GetSubject()
			if err != nil || sub == "" {
				sendUnauthorized(w, "Token is missing subject (sub) claim")
				return
			}

			ctx := context.WithValue(r.Context(), constants.UserContextKey{}, sub)
			r = r.WithContext(ctx)

			r.Header.Set("X-User-Id", sub)

			next.ServeHTTP(w, r)
		})
	}
}

// sendUnauthorized writes a 401 JSON response.
func sendUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"code":"unauthorized","message":%q}`, message)
}

func RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Context().Value(constants.UserContextKey{}) == nil {
				sendUnauthorized(w, "Missing authentication")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ExtractUserID(r *http.Request) *string {
	return r.Context().Value(constants.UserContextKey{}).(*string)
}
