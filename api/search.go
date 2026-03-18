package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func (b *ProductsBinder) searchHandler(c *gin.Context) {
	// Get the search query from Unbxd URL 
	query, _ := url.QueryUnescape(c.Param("query"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
		return
	}

	// Build Unbxd API URL with search query's base URL
	baseURL := "https://search.unbxd.io/1ccbb7fcb0faf770d1c228be80ba16d9/ss-unbxd-aus-demo-fashion831421736321881/search"
	apiURL := baseURL + "?q=" + url.QueryEscape(query)

	// Calling Unbxd search API
	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Search service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Response status
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Search service error"})
		return
	}

	// Read response from Unbxd API
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse the results"})
		return
	}

	// Return the search results
	c.JSON(http.StatusOK, result)
}
