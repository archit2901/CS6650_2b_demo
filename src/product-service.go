package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// Product mirrors components/schemas/Product in api.yaml.
type Product struct {
	ProductID    int32  `json:"product_id" binding:"required,min=1"`
	SKU          string `json:"sku" binding:"required,min=1,max=100"`
	Manufacturer string `json:"manufacturer" binding:"required,min=1,max=200"`
	CategoryID   int32  `json:"category_id" binding:"required,min=1"`
	Weight       *int32 `json:"weight" binding:"required,min=0"`
	SomeOtherID  int32  `json:"some_other_id" binding:"required,min=1"`
}

// ErrorResponse mirrors components/schemas/Error.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ProductStore is our in-memory "database".
type ProductStore struct {
	mu       sync.RWMutex
	products map[int32]Product
}

// NewProductStore pre-loads products 1..n, since the API has no create endpoint.
func NewProductStore(n int) *ProductStore {
	s := &ProductStore{products: make(map[int32]Product)}
	for i := int32(1); i <= int32(n); i++ {
		w := int32(1000)
		s.products[i] = Product{
			ProductID: i, SKU: fmt.Sprintf("SKU-%d", i), Manufacturer: "Acme Corporation",
			CategoryID: 1, Weight: &w, SomeOtherID: 1,
		}
	}
	return s
}

func (s *ProductStore) Get(id int32) (Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.products[id]
	return p, ok
}

// Update replaces details of an existing product. Returns false if it doesn't exist.
func (s *ProductStore) Update(p Product) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.products[p.ProductID]; !ok {
		return false
	}
	s.products[p.ProductID] = p
	return true
}

// sendError writes a response shaped like the spec's Error schema.
func sendError(c *gin.Context, status int, code, message, details string) {
	c.JSON(status, ErrorResponse{Error: code, Message: message, Details: details})
}

// GET /products/{productId}  ->  200, 404, 500
func getProductHandler(store *ProductStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("productId"), 10, 32)
		if err != nil || id < 1 {
			// GET has no 400 in the spec; an invalid ID can never match a product.
			sendError(c, http.StatusNotFound, "NOT_FOUND", "Product not found", "productId must be a positive integer")
			return
		}
		p, ok := store.Get(int32(id))
		if !ok {
			sendError(c, http.StatusNotFound, "NOT_FOUND", "Product not found", fmt.Sprintf("no product with id %d", id))
			return
		}
		c.JSON(http.StatusOK, p)
	}
}

// POST /products/{productId}/details  ->  204, 400, 404, 500
func addProductDetailsHandler(store *ProductStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("productId"), 10, 32)
		if err != nil || id < 1 {
			sendError(c, http.StatusBadRequest, "INVALID_INPUT", "The provided input data is invalid", "productId must be a positive integer")
			return
		}
		var p Product
		if err := c.ShouldBindJSON(&p); err != nil {
			sendError(c, http.StatusBadRequest, "INVALID_INPUT", "The provided input data is invalid", err.Error())
			return
		}
		if int64(p.ProductID) != id {
			sendError(c, http.StatusBadRequest, "INVALID_INPUT", "The provided input data is invalid", "product_id in body must match productId in path")
			return
		}
		if !store.Update(p) {
			sendError(c, http.StatusNotFound, "NOT_FOUND", "Product not found", fmt.Sprintf("no product with id %d", id))
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// recoveryHandler turns any panic into a spec-shaped 500 instead of crashing.
func recoveryHandler(c *gin.Context, recovered any) {
	sendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "")
	c.Abort()
}

// faultInjection lets us deliberately trigger a 500 for testing.
func faultInjection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Simulate-Error") == "true" {
			panic("simulated internal failure")
		}
		c.Next()
	}
}

func main() {
	store := NewProductStore(100)

	router := gin.New()
	router.Use(gin.Logger(), gin.CustomRecovery(recoveryHandler), faultInjection())

	router.GET("/products/:productId", getProductHandler(store))
	router.POST("/products/:productId/details", addProductDetailsHandler(store))

	router.Run(":8080")
}
