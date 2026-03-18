package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	httptransport "github.com/unbxd/go-base/kit/transport/http"
	"github.com/unbxd/go-base/utils/log"
	_ "github.com/lib/pq"
)

type ProductsBinder struct {
	logger log.Logger
	db     *sql.DB
	rdb    *redis.Client
	ctx    context.Context
}

func NewProductsBinder(logger log.Logger) (*ProductsBinder, error) {
	envPaths := []string{
		".env",                           // Current directory
		filepath.Join("..", ".env"),      // Parent directory (when running from cmd/app)
		filepath.Join("../..", ".env"),   // Two levels up
		filepath.Join(".", ".env"),       // Explicit current directory
	}
	
	var envLoaded bool
	for _, envPath := range envPaths {
		if err := godotenv.Load(envPath); err == nil {
			logger.Info("Loaded .env file", log.String("path", envPath))
			envLoaded = true
			break
		}
	}
	
	if !envLoaded {
		logger.Info("Warning: Could not load .env file, using system environment variables")
	}

	// Initialize database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "my_db")

	var connStr string
	if dbPassword != "" {
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, dbName)
	} else {
		connStr = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbName)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping database")
	}

	logger.Info("Connected to database")

	// Initialize Redis connection
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDBStr := getEnv("REDIS_DB", "0")
	redisDB := 0
	if dbNum, err := strconv.Atoi(redisDBStr); err == nil {
		redisDB = dbNum
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       redisDB,
	})

	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		logger.Info("Warning: Redis not available - caching will be disabled")
		rdb = nil
	} else {
		logger.Info("Connected to Redis")
	}

	return &ProductsBinder{
		logger: logger,
		db:     db,
		rdb:    rdb,
		ctx:    ctx,
	}, nil
}

func (b *ProductsBinder) Bind(ht *httptransport.Transport, opts ...httptransport.HandlerOption) {
	// Create Gin router
	r := gin.Default()

	// Register all routes with service-based paths
	r.GET("/search/:query", b.searchHandler)
	r.POST("/ingest", b.ingestHandler)
	r.GET("/browse/category/:catlevel1Name/:catlevel2Name/:catlevel3Name/:categoryType", b.getProductsHandler)
	r.GET("/browse/:id", b.getProductHandler)
	r.DELETE("/delete_prod/:id", b.deleteProductHandler)
	r.DELETE("/delete_prod", b.deleteProductsHandler)

	// Mount Gin router to HTTP transport
	// Register for all HTTP methods - Gin router handles method routing internally
	// Register GET first, then other methods
	ht.Mux().Handler(http.MethodGet, "/*", r)
	ht.Mux().Handler(http.MethodPost, "/*", r)
	ht.Mux().Handler(http.MethodDelete, "/*", r)
	ht.Mux().Handler(http.MethodPut, "/*", r)
	ht.Mux().Handler(http.MethodPatch, "/*", r)
	ht.Mux().Handler(http.MethodOptions, "/*", r)
	ht.Mux().Handler(http.MethodHead, "/*", r)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
