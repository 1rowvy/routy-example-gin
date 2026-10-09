package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthRequired lets through requests with a valid `Authorization: Bearer <token>`.
func AuthRequired(s *store) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		user, ok := s.userByToken(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

// SessionRequired lets through requests with a valid `session` cookie.
func SessionRequired(s *store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := c.Cookie("session")
		user, ok := s.userBySession(id)
		if err != nil || !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no session"})
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

// AdminOnly must run after AuthRequired.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if currentUser(c).Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admins only"})
			return
		}
		c.Next()
	}
}
