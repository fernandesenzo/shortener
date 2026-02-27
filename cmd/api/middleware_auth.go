package main

import (
	"net/http"
	"strings"

	"github.com/fernandesenzo/shortener/internal/identity"
	"github.com/fernandesenzo/shortener/internal/jwt"
)

func AuthMiddleware(next http.Handler, jwtManager *jwt.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := ""
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			parsedID, err := jwtManager.ValidateToken(tokenString)
			if err == nil {
				userID = parsedID
			}
		}
		ctx := identity.WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
