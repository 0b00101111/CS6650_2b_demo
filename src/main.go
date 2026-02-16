package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// Product matches the OpenAPI schema exactly
type Product struct {
	ProductID    int    `json:"product_id" binding:"required,min=1"`
	SKU          string `json:"sku" binding:"required,min=1,max=100"`
	Manufacturer string `json:"manufacturer" binding:"required,min=1,max=200"`
	CategoryID   int    `json:"category_id" binding:"required,min=1"`
	Weight       int    `json:"weight" binding:"min=0"`
	SomeOtherID  int    `json:"some_other_id" binding:"required,min=1"`
}

// Error response matches the OpenAPI Error schema
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// In-memory storage using sync.Map for thread safety
var (
	products = make(map[int]Product)
	mu       sync.RWMutex
)

func main() {
	router := gin.Default()

	router.GET("/products/:productId", getProduct)
	router.POST("/products/:productId/details", addProductDetails)

	router.Run(":8080")
}

// GET /products/{productId} - Retrieve a product by ID
func getProduct(c *gin.Context) {
	// Parse and validate productId from path
	productId, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productId < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid product ID",
			Details: "Product ID must be a positive integer",
		})
		return
	}

	// Look up product in memory
	mu.RLock()
	product, exists := products[productId]
	mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "NOT_FOUND",
			Message: "Product not found",
			Details: "No product found with ID " + strconv.Itoa(productId),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

// POST /products/{productId}/details - Add or update product details
func addProductDetails(c *gin.Context) {
	// Parse and validate productId from path
	productId, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productId < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid product ID",
			Details: "Product ID must be a positive integer",
		})
		return
	}

	// Bind and validate request body
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid input data",
			Details: err.Error(),
		})
		return
	}

	// Store the product
	mu.Lock()
	products[productId] = product
	mu.Unlock()

	c.Status(http.StatusNoContent) // 204
}