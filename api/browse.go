package api

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ProductsResponse struct {
	Data     []map[string]interface{} `json:"data"`
	Page     int                       `json:"page"`
	PerPage  int                       `json:"perPage"`
	Total    int                       `json:"total"`
}

type ProductResponse struct {
	Data map[string]interface{} `json:"data"`
}

const (
	cacheTTLProduct  = 1 * time.Hour
	cacheTTLProducts = 1 * time.Hour
)

// Browse single product by ID
func (b *ProductsBinder) getProductHandler(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID required"})
		return
	}

	// Check redis cache first
	cacheKey := fmt.Sprintf("product:%s", productID)
	if b.rdb != nil {
		cachedData, err := b.rdb.Get(b.ctx, cacheKey).Result()
		if err == nil {
			var item map[string]interface{}
			json.Unmarshal([]byte(cachedData), &item)
			c.JSON(http.StatusOK, ProductResponse{Data: item})
			return
		}
	}

	// Get from postgresql database
	query := "SELECT unique_id, sku, name, title, product_description, product_url, product_image, price, availability, category_type, catlevel1_name, catlevel2_name, catlevel3_name, catlevel4_name, color, size, category, gender FROM catalog_data WHERE unique_id = $1"
	
	var uniqueID, sku, name, title, productDesc, productURL, productImage, availability, categoryType string
	var catlevel1, catlevel2, catlevel3, catlevel4 sql.NullString
	var price sql.NullFloat64
	var colorJSON, sizeJSON, categoryJSON, genderJSON string

	err := b.db.QueryRow(query, productID).Scan(&uniqueID, &sku, &name, &title, &productDesc, &productURL, &productImage, &price, &availability, &categoryType, &catlevel1, &catlevel2, &catlevel3, &catlevel4, &colorJSON, &sizeJSON, &categoryJSON, &genderJSON)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// JSON strings to arrays
	var color, size, category, gender []interface{}
	json.Unmarshal([]byte(colorJSON), &color)
	json.Unmarshal([]byte(sizeJSON), &size)
	json.Unmarshal([]byte(categoryJSON), &category)
	json.Unmarshal([]byte(genderJSON), &gender)

	// Build the response structure
	item := map[string]interface{}{
		"uniqueId":          uniqueID,
		"sku":               sku,
		"name":              name,
		"title":             title,
		"productDescription": productDesc,
		"productUrl":        productURL,
		"productImage":      productImage,
		"price":             price.Float64,
		"availability":       availability,
		"categoryType":      categoryType,
		"catlevel1Name":     catlevel1.String,
		"catlevel2Name":     catlevel2.String,
		"catlevel3Name":     catlevel3.String,
		"catlevel4Name":     catlevel4.String,
		"color":             color,
		"size":              size,
		"category":          category,
		"gender":            gender,
	}

	// Save to cache (as it is not present in cache already)
	if b.rdb != nil {
		itemJSON, _ := json.Marshal(item)
		b.rdb.Set(b.ctx, cacheKey, itemJSON, cacheTTLProduct)
	}

	c.JSON(http.StatusOK, ProductResponse{Data: item})
}

// Browse category-level list of products with optionalfilters
// Path level parameters: /products/category/:catlevel1Name/:catlevel2Name/:catlevel3Name/:categoryType
// Query level parameters: page, sort, minPrice, maxPrice
func (b *ProductsBinder) getProductsHandler(c *gin.Context) {
	// Get path level parameters
	catLevel1, _ := url.QueryUnescape(c.Param("catlevel1Name"))
	catLevel2, _ := url.QueryUnescape(c.Param("catlevel2Name"))
	catLevel3, _ := url.QueryUnescape(c.Param("catlevel3Name"))
	categoryType, _ := url.QueryUnescape(c.Param("categoryType"))
	
	// Use "-"" as placeholder for empty values
	if catLevel1 == "-" {
		catLevel1 = ""
	}
	if catLevel2 == "-" {
		catLevel2 = ""
	}
	if catLevel3 == "-" {
		catLevel3 = ""
	}
	if categoryType == "-" {
		categoryType = ""
	}
	
	// Get the query level parameters
	minPriceStr := c.Query("minPrice")
	maxPriceStr := c.Query("maxPrice")
	sortOrder := c.Query("sort")
	pageStr := c.Query("page")
	
	// Set default values for optional parameters
	if sortOrder == "" {
		sortOrder = "DESC" // Default sorting to descending order
	}
	sortOrder = strings.ToUpper(sortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC" 
	}
	
	page := 1 // Default shows page 1
	p, err := strconv.Atoi(pageStr)
	if err == nil && p > 0 {
    	page = p
	}
	perPage := 10
	offset := (page - 1) * perPage

	// Check cache
	cacheKey := generateCacheKey(catLevel1, catLevel2, catLevel3, categoryType, minPriceStr, maxPriceStr, sortOrder, page, perPage)
	if b.rdb != nil {
		cachedData, err := b.rdb.Get(b.ctx, cacheKey).Result()
		if err == nil {
			var response ProductsResponse
			json.Unmarshal([]byte(cachedData), &response)
			c.JSON(http.StatusOK, response)
			return
		}
	}

	// Build WHERE clause for optional filters
	whereClause := "WHERE 1=1"
	var args []interface{}
	argNum := 1

	// Add optional filters
	if catLevel1 != "" {
		whereClause += fmt.Sprintf(" AND catlevel1_name = $%d", argNum)
		args = append(args, catLevel1)
		argNum++
	}
	if catLevel2 != "" {
		whereClause += fmt.Sprintf(" AND catlevel2_name = $%d", argNum)
		args = append(args, catLevel2)
		argNum++
	}
	if catLevel3 != "" {
		whereClause += fmt.Sprintf(" AND catlevel3_name = $%d", argNum)
		args = append(args, catLevel3)
		argNum++
	}
	if categoryType != "" {
		whereClause += fmt.Sprintf(" AND category_type = $%d", argNum)
		args = append(args, categoryType)
		argNum++
	}
	if minPriceStr != "" {
		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err == nil {
			whereClause += fmt.Sprintf(" AND price >= $%d", argNum)
			args = append(args, minPrice)
			argNum++
		}
	}
	if maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err == nil {
			whereClause += fmt.Sprintf(" AND price <= $%d", argNum)
			args = append(args, maxPrice)
			argNum++
		}
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM catalog_data " + whereClause
	var total int
	b.db.QueryRow(countQuery, args...).Scan(&total)

	// Build main query
	query := "SELECT unique_id, sku, name, title, product_description, product_url, product_image, price, availability, category_type, catlevel1_name, catlevel2_name, catlevel3_name, catlevel4_name, color, size, category, gender FROM catalog_data " + whereClause
	query += fmt.Sprintf(" ORDER BY price %s LIMIT $%d OFFSET $%d", sortOrder, argNum, argNum+1)
	args = append(args, perPage, offset)

	// Execute the postgresql query
	rows, err := b.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	// Read the results
	results := []map[string]interface{}{}
	for rows.Next() {
		var uniqueID, sku, name, title, productDesc, productURL, productImage, availability, categoryType string
		var catlevel1, catlevel2, catlevel3, catlevel4 sql.NullString
		var price sql.NullFloat64
		var colorJSON, sizeJSON, categoryJSON, genderJSON string

		rows.Scan(&uniqueID, &sku, &name, &title, &productDesc, &productURL, &productImage, &price, &availability, &categoryType, &catlevel1, &catlevel2, &catlevel3, &catlevel4, &colorJSON, &sizeJSON, &categoryJSON, &genderJSON)

		// Convert JSON strings to arrays
		var color, size, category, gender []interface{}
		json.Unmarshal([]byte(colorJSON), &color)
		json.Unmarshal([]byte(sizeJSON), &size)
		json.Unmarshal([]byte(categoryJSON), &category)
		json.Unmarshal([]byte(genderJSON), &gender)

		item := map[string]interface{}{
			"uniqueId":uniqueID,
			"sku":sku,
			"name":name,
			"title":title,
			"productDescription":productDesc,
			"productUrl":productURL,
			"productImage":productImage,
			"price":price.Float64,
			"availability":availability,
			"categoryType":categoryType,
			"catlevel1Name":catlevel1.String,
			"catlevel2Name":catlevel2.String,
			"catlevel3Name":catlevel3.String,
			"catlevel4Name":catlevel4.String,
			"color":color,
			"size":size,
			"category":category,
			"gender":gender,
		}

		results = append(results, item)
	}

	// Build the response structure
	response := ProductsResponse{
		Data:    results,
		Page:    page,
		PerPage: perPage,
		Total:   total,
	}

	// Save to cache
	if b.rdb != nil {
		responseJSON, _ := json.Marshal(response)
		b.rdb.Set(b.ctx, cacheKey, responseJSON, cacheTTLProducts)
	}

	c.JSON(http.StatusOK, response)
}

// Generate the cache key from all the parameters
func generateCacheKey(catLevel1, catLevel2, catLevel3, categoryType, minPrice, maxPrice, sortOrder string, page, perPage int) string {
	key := fmt.Sprintf("products|%s|%s|%s|%s|%s|%s|%s|%d|%d",
		catLevel1, catLevel2, catLevel3, categoryType, minPrice, maxPrice, sortOrder, page, perPage)
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("products:%x", hash)
}
