package api

import (
	"context"
	"log"
	"net/http"
	"new-billing/internal/config"
	"new-billing/internal/models"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const UserContextKey = contextKey("user_claims")

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.Auth.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleRequired(allowedRoles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*jwt.RegisteredClaims)
			if !ok {
				http.Error(w, "User claims not found in context", http.StatusInternalServerError)
				return
			}

			userRole := models.Role(claims.Audience[0])
			isAllowed := false
			for _, role := range allowedRoles {
				// Админ имеет доступ ко всему, к чему имеет доступ менеджер
				if role == models.ManagerRole && userRole == models.AdminRole {
					isAllowed = true
					break
				}
				if userRole == role {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				log.Printf("Forbidden: User with role '%s' tried to access a resource requiring one of %v", userRole, allowedRoles)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
