// Package api wires the HTTP routes of the shop.
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/1rowvy/outry-example-gin/internal/model"
)

const version = "1.0.0"

// Router builds the gin engine with every route of the service.
func Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	s := newStore()

	r.GET("/health", Health)

	auth := r.Group("/auth")
	auth.POST("/login", s.Login)
	auth.POST("/session", s.StartSession)
	auth.GET("/me", SessionRequired(s), s.Me)

	v1 := r.Group("/v1", AuthRequired(s))
	registerUsers(v1, s)
	registerOrders(v1.Group("/orders"), s)

	return r
}

// Health
// Liveness probe for load balancers.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, model.Health{Status: "ok", Version: version})
}
