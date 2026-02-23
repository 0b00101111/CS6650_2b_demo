package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
}

type SearchResponse struct {
	Products   []Product `json:"products"`
	TotalFound int       `json:"total_found"`
	SearchTime string    `json:"search_time"`
}

// In-memory product store
var products []Product

// Sample data for generation
var brands = []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta", "Iota", "Kappa"}
var categories = []string{"Electronics", "Books", "Home", "Sports", "Clothing", "Food", "Toys", "Health", "Garden", "Auto"}
var descriptions = []string{
	"High quality product with excellent features",
	"Best seller in its category",
	"Premium grade with warranty included",
	"Affordable and reliable choice",
	"Top rated by customers worldwide",
}

func generateProducts(count int) []Product {
	prods := make([]Product, count)
	for i := 0; i < count; i++ {
		brand := brands[i%len(brands)]
		prods[i] = Product{
			ID:          i + 1,
			Name:        fmt.Sprintf("Product %s %d", brand, i+1),
			Category:    categories[i%len(categories)],
			Description: descriptions[i%len(descriptions)],
			Brand:       brand,
		}
	}
	return prods
}

func searchProducts(query string) SearchResponse {
	start := time.Now()
	query = strings.ToLower(query)

	var results []Product
	totalFound := 0
	checked := 0

	// Check exactly 100 products then stop
	for _, p := range products {
		if checked >= 100 {
			break
		}
		checked++

		// Case-insensitive search on name and category
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Category), query) {
			totalFound++
			if len(results) < 20 {
				results = append(results, p)
			}
		}
	}

	elapsed := time.Since(start)

	if results == nil {
		results = []Product{}
	}

	return SearchResponse{
		Products:   results,
		TotalFound: totalFound,
		SearchTime: elapsed.String(),
	}
}

func main() {
	// Generate 100,000 products at startup
	fmt.Println("Generating 100,000 products...")
	products = generateProducts(100000)
	fmt.Printf("Generated %d products\n", len(products))

	router := gin.Default()

	// Health check endpoint (for ALB in Part 3)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Search endpoint
	router.GET("/products/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "MISSING_QUERY",
				"message": "Search query parameter 'q' is required",
			})
			return
		}
		result := searchProducts(query)
		c.JSON(http.StatusOK, result)
	})

	router.Run(":8080")
}
