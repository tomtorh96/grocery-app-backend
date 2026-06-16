package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(getEnv("JWT_SECRET", "dev-secret-change-in-production"))

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plain text password against a hash
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateToken creates a JWT token for a user
func GenerateToken(userID string, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken parses and validates a JWT token, returns userID and username
func ValidateToken(tokenStr string) (string, string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid claims")
	}

	userID, _ := claims["user_id"].(string)
	username, _ := claims["username"].(string)
	return userID, username, nil
}

// Middleware extracts and validates JWT from Authorization header
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) < 8 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr := authHeader[7:] // strip "Bearer "
		userID, username, err := ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
		ctx = context.WithValue(ctx, contextKeyUsername, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// context keys
type contextKey string

const (
	contextKeyUserID   contextKey = "user_id"
	contextKeyUsername contextKey = "username"
)

// GetUserID extracts the user ID from the request context
func GetUserID(r *http.Request) string {
	id, _ := r.Context().Value(contextKeyUserID).(string)
	return id
}

// GetUsername extracts the username from the request context
func GetUsername(r *http.Request) string {
	name, _ := r.Context().Value(contextKeyUsername).(string)
	return name
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
