package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// secretKey is used to verify the JWT signature, loaded from an environment variable.
var secretKey = os.Getenv("RS256PK")

// contextKey is a type alias for string, used for defining context keys in a type-safe way.
type contextKey string

// ClaimsKey is a constant key used to store JWT claims in the request context.
const ClaimsKey contextKey = "claims"

// AuthMiddleware is a JWT authentication middleware that validates the token and extracts claims.
// It intercepts requests to check for a valid JWT in the Authorization header, and adds claims to the request context.
func AuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token mancante", http.StatusUnauthorized) // Return 401 if the token is missing
			return
		}

		// Split the header into "Bearer" and token parts
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Formato del token non valido", http.StatusUnauthorized) // Return 401 if the format is invalid
			return
		}

		tokenString := parts[1] // Extract the actual JWT

		// Parse and validate the JWT using the secret key
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the token is signed with the correct method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Metodo di firma non valido: %v", token.Header["alg"]) // Return error if signing method is invalid
			}
			return []byte(secretKey), nil // Return the secret key for signature verification
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token non valido", http.StatusUnauthorized) // Return 401 if the token is invalid
			return
		}

		// Extract claims from the token and verify they are in the expected format
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Claims non validi", http.StatusUnauthorized) // Return 401 if claims are not valid
			return
		}

		// Add claims to the request context for use in downstream handlers
		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		r = r.WithContext(ctx)

		// Pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}
