package middleware

import (
	"context"
	"net/http"
	"strings"

	"order-service/internal/service"

	"github.com/gin-gonic/gin"
)

type contextKey string

const tokenKey contextKey = "token"
const userIDKey contextKey = "userID"

var userClient *service.UserClient

// InitUserClient initializes the user service client
func InitUserClient(userServiceURL string) {
	userClient = service.NewUserClient(userServiceURL)
}

// ExtractToken extracts JWT token from Authorization header and validates with User Service
func ExtractToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token with User Service
		valid, userID, err := userClient.ValidateToken(c.Request.Context(), token)
		if err != nil || !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store token and userID in request context
		ctx := context.WithValue(c.Request.Context(), tokenKey, token)
		ctx = context.WithValue(ctx, userIDKey, userID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetToken retrieves token from context
func GetToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey).(string)
	return token, ok
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
