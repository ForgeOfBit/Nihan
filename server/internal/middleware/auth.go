package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// AuthUserIDKey is the context key where the authenticated user's ID is stored.
	AuthUserIDKey = "auth_user_id"
)

// JWTClaims represents the claims embedded in an access token.
type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthMiddleware returns a Gin middleware that validates Bearer JWT tokens.
// It extracts the user ID from the token and stores it in the Gin context.
func AuthMiddleware(accessSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header must be in the format: Bearer <token>",
			})
			return
		}

		tokenString := parts[1]

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is HMAC.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(accessSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user ID in token",
			})
			return
		}

		c.Set(AuthUserIDKey, userID)
		c.Next()
	}
}

// GetUserID is a helper that retrieves the authenticated user's UUID from
// the Gin context. It panics if called outside of an authenticated route.
func GetUserID(c *gin.Context) uuid.UUID {
	val, exists := c.Get(AuthUserIDKey)
	if !exists {
		// This should never happen if the middleware is applied correctly.
		panic("middleware: GetUserID called without AuthMiddleware")
	}
	return val.(uuid.UUID)
}
