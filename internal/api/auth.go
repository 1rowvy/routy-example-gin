package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/1rowvy/routy-example-gin/internal/model"
)

// Login
// Exchanges email and password for a bearer token that lives for an hour.
func (s *store) Login(c *gin.Context) {
	var req model.Login
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, ok := s.checkPassword(req.Email, req.Password)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong email or password"})
		return
	}
	c.JSON(http.StatusOK, model.Token{Token: s.newToken(user.ID), ExpiresIn: 3600})
}

// Start session
// Browser-style login: a form post that sets the `session` cookie.
func (s *store) StartSession(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	user, ok := s.checkPassword(email, password)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong email or password"})
		return
	}
	c.SetCookie("session", s.newSession(user.ID), 3600, "/", "", false, true)
	c.Status(http.StatusNoContent)
}

// Me
// The user of the current session.
func (s *store) Me(c *gin.Context) {
	c.JSON(http.StatusOK, currentUser(c))
}
