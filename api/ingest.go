package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unbxd/go-base/utils/log"
)

type CatalogItem struct {
	Color              []string `json:"color"`
	CategoryType       string   `json:"categoryType"`
	ProductURL         string   `json:"productUrl"`
	Availability       string   `json:"availability"`
	Size               []string `json:"size"`
	Category           []string `json:"category"`
	ProductDescription string   `json:"productDescription"`
	Catlevel2Name      string   `json:"catlevel2Name"`
	Title              string   `json:"title"`
	ProductImage       string   `json:"productImage"`
	SKU                string   `json:"sku"`
	Price              float64  `json:"price"`
	Catlevel3Name      string   `json:"catlevel3Name"`
	Catlevel1Name      string   `json:"catlevel1Name"`
	Name               string   `json:"name"`
	Gender             []string `json:"gender"`
	Catlevel4Name      string   `json:"catlevel4Name"`
	UniqueID           string   `json:"uniqueId"`
}

type IngestResponse struct {
	Inserted int    `json:"inserted"`
	Errors   int    `json:"errors"`
	Message  string `json:"message"`
}

func (b *ProductsBinder) ingestHandler(c *gin.Context) {
	// Read request body
	bodyBytes, _ := io.ReadAll(c.Request.Body)

	var items []CatalogItem
	
	// Parse as array first
	json.Unmarshal(bodyBytes, &items)
	
	// If not an array, try as single object
	if len(items) == 0 {
		var singleItem CatalogItem
		json.Unmarshal(bodyBytes, &singleItem)
		if singleItem.UniqueID != "" {
			items = []CatalogItem{singleItem}
		}
	}

	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	b.logger.Info("Processing ingest batch", log.String("count", strconv.Itoa(len(items))))

	inserted := 0
	errors := 0

	// Insert each product into the DB
	for i, item := range items {
		// Convert arrays to JSON strings
		colorJSON, _ := json.Marshal(item.Color)
		sizeJSON, _ := json.Marshal(item.Size)
		categoryJSON, _ := json.Marshal(item.Category)
		genderJSON, _ := json.Marshal(item.Gender)

		// SQL query to insert or update product
		query := `INSERT INTO catalog_data (
			unique_id, sku, name, title, product_description, product_url,
			product_image, price, availability, category_type,
			catlevel1_name, catlevel2_name, catlevel3_name, catlevel4_name,
			color, size, category, gender
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (unique_id) DO UPDATE SET
			sku = EXCLUDED.sku,
			name = EXCLUDED.name,
			title = EXCLUDED.title,
			product_description = EXCLUDED.product_description,
			product_url = EXCLUDED.product_url,
			product_image = EXCLUDED.product_image,
			price = EXCLUDED.price,
			availability = EXCLUDED.availability,
			category_type = EXCLUDED.category_type,
			catlevel1_name = EXCLUDED.catlevel1_name,
			catlevel2_name = EXCLUDED.catlevel2_name,
			catlevel3_name = EXCLUDED.catlevel3_name,
			catlevel4_name = EXCLUDED.catlevel4_name,
			color = EXCLUDED.color,
			size = EXCLUDED.size,
			category = EXCLUDED.category,
			gender = EXCLUDED.gender`

		_, err := b.db.Exec(query,
			item.UniqueID, item.SKU, item.Name, item.Title, item.ProductDescription,
			item.ProductURL, item.ProductImage, item.Price, item.Availability,
			item.CategoryType, item.Catlevel1Name, item.Catlevel2Name,
			item.Catlevel3Name, item.Catlevel4Name,
			string(colorJSON), string(sizeJSON), string(categoryJSON), string(genderJSON))

		if err != nil {
			b.logger.Error("Ingest item failed", log.String("index", strconv.Itoa(i+1)), log.Error(err))
			errors++
		} else {
			inserted++
		}
	}

	// Return response
	statusCode := http.StatusCreated
	if errors > 0 && inserted == 0 {
		statusCode = http.StatusBadRequest
	} else if errors > 0 {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, IngestResponse{
		Inserted: inserted,
		Errors:   errors,
		Message:  "Ingestion completed",
	})
}
