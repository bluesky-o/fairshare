package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/bluesky-o/fairshare/pkg/firebase"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Authenticate(firebaseClient *firebase.Client) func (http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeUnauthorized(w, "invalid authorization header format")
				return
			}

			idToken := parts[1]
			token, err := firebaseClient.VerifyToken(r.Context(), idToken)
			if err != nil {
				writeUnauthorized(w, "invalid or expired token")
			}

			ctx := context.WithValue(r.Context(), UserIDKey, token.UID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized) 
	w.Write([]byte(`{"error":"unauthorized","message":"` + message + `"}`))
}
