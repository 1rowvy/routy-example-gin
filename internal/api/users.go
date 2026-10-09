package api

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/1rowvy/routy-example-gin/internal/model"
)

func registerUsers(r *gin.RouterGroup, s *store) {
	r.POST("/users", s.CreateUser)
	r.GET("/users/:id", s.GetUser)
	r.POST("/users/:id/avatar", s.UploadAvatar)
	r.DELETE("/users/:id", AdminOnly(), s.DeleteUser)
}

// Create user
// The email must be unique; the role defaults to "customer".
func (s *store) CreateUser(c *gin.Context) {
	var req model.CreateUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := s.createUser(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

// Get user
func (s *store) GetUser(c *gin.Context) {
	user, ok := s.user(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Upload avatar
// Multipart form with an image in the `file` field.
func (s *store) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	user, ok := s.setAvatar(c.Param("id"), "/avatars/"+filepath.Base(file.Filename))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Delete user
// Admins only.
func (s *store) DeleteUser(c *gin.Context) {
	if !s.deleteUser(c.Param("id")) {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such user"})
		return
	}
	c.Status(http.StatusNoContent)
}
