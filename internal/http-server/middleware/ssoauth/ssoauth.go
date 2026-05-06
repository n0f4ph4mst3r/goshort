package ssoauth

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	AppID  int64  `json:"app_id"`
	jwt.RegisteredClaims
}

func New(log *slog.Logger, publicKeyPath string) func(http.Handler) http.Handler {
	keyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Error("failed to read public key: " + err.Error())
		os.Exit(1)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		log.Error("failed to parse public key: " + err.Error())
		os.Exit(1)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid auth header", http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]
			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return publicKey, nil
			})

			if err != nil || !token.Valid {
				log.Warn("invalid token", slog.String("err", err.Error()))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
