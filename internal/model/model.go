// Package model holds the types the API reads and writes.
package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Avatar    *string   `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUser struct {
	Name  string `json:"name" binding:"required"`  // full name
	Email string `json:"email" binding:"required"` // must be unique
	Role  string `json:"role"`                     // "admin" or "customer" (default)
}

type Item struct {
	SKU   string  `json:"sku" binding:"required"`
	Qty   int     `json:"qty" binding:"required"`
	Price float64 `json:"price"`
}

type Order struct {
	ID      int        `json:"id"`
	UserID  int        `json:"user_id"`
	Items   []Item     `json:"items"`
	Total   float64    `json:"total"`
	Status  string     `json:"status"` // new → pending → paid
	Note    string     `json:"note,omitempty"`
	PaidAt  *time.Time `json:"paid_at"`
	Created time.Time  `json:"created_at"`
}

type CreateOrder struct {
	UserID int    `json:"user_id" binding:"required"`
	Items  []Item `json:"items" binding:"required"`
	Note   string `json:"note"`
}

type OrderPage struct {
	Items []Order `json:"items"`
	Page  int     `json:"page"`
	Total int     `json:"total"`
}

type Login struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Token struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

type Health struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
