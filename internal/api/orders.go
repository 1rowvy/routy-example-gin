package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/1rowvy/routy-example-gin/internal/model"
)

func registerOrders(r *gin.RouterGroup, s *store) {
	r.GET("", s.ListOrders)
	r.POST("", s.CreateOrder)
	r.GET("/:id", s.GetOrder)
	r.POST("/:id/pay", s.PayOrder)
}

// List orders
// Orders of the current user, newest first, 20 per page.
func (s *store) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	status := c.Query("status")
	c.JSON(http.StatusOK, s.listOrders(currentUser(c).ID, status, page))
}

// Create order
// Prices come from the catalog; `X-Idempotency-Key` makes retries safe.
func (s *store) CreateOrder(c *gin.Context) {
	var req model.CreateOrder
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := s.createOrder(req, c.GetHeader("X-Idempotency-Key"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// Get order
func (s *store) GetOrder(c *gin.Context) {
	order, ok := s.order(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such order"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// Pay order
// Payment is asynchronous: the order becomes "pending" and turns "paid" a moment later.
func (s *store) PayOrder(c *gin.Context) {
	order, err := s.pay(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, order)
}
