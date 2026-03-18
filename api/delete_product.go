package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unbxd/go-base/utils/log"
)

type DeleteRequest struct {
	Ids []string `json:"ids"`
}

type DeleteResponse struct {
	Deleted int    `json:"deleted"`
	Errors  int    `json:"errors"`
	Message string `json:"message"`
}

// Delete single product by ID
func (b *ProductsBinder) deleteProductHandler(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID required"})
		return
	}

	// Delete from postgresql database
	query := "DELETE FROM catalog_data WHERE unique_id = $1"
	result, err := b.db.Exec(query, productID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if product was deleted
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Delete product from Redis cache if it exists
	if b.rdb != nil {
		cacheKey := fmt.Sprintf("product:%s", productID)
		b.rdb.Del(b.ctx, cacheKey)
		
		// This deletes all product list cache
		iter := b.rdb.Scan(b.ctx, 0, "products:*", 0).Iterator()
		for iter.Next(b.ctx) {
			b.rdb.Del(b.ctx, iter.Val())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
		"deleted": 1,
	})
}

// Delete multiple products by IDs
func (b *ProductsBinder) deleteProductsHandler(c *gin.Context) {
	// Read request body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var deleteReq DeleteRequest
	err = json.Unmarshal(bodyBytes, &deleteReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Check if IDs array is empty
	if len(deleteReq.Ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No product IDs provided"})
		return
	}

	deleted := 0
	errors := 0

	// Delete each product
	for _, productID := range deleteReq.Ids {
		query := "DELETE FROM catalog_data WHERE unique_id = $1"
		result, err := b.db.Exec(query, productID)

		if err != nil {
			b.logger.Error("Error deleting product", log.String("productId", productID), log.Error(err))
			errors++
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			deleted++
			
			// Delete from Redis cache if it exists
			if b.rdb != nil {
				cacheKey := fmt.Sprintf("product:%s", productID)
				b.rdb.Del(b.ctx, cacheKey)
			}
		} else {
			errors++
		}
	}

	if b.rdb != nil && deleted > 0 {
		// This deletes all product list cache
		iter := b.rdb.Scan(b.ctx, 0, "products:*", 0).Iterator()
		for iter.Next(b.ctx) {
			b.rdb.Del(b.ctx, iter.Val())
		}
	}

	// Return response
	statusCode := http.StatusOK
	if deleted == 0 {
		statusCode = http.StatusNotFound
	}

	c.JSON(statusCode, DeleteResponse{
		Deleted: deleted,
		Errors:  errors,
		Message: "Deletion completed",
	})
}