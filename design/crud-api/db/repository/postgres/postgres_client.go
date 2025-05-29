package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"lk/datafoundation/crud-api/db/config"

	_ "github.com/lib/pq"
)

// Config holds the database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Client represents a PostgreSQL database client
type Client struct {
	db *sql.DB
}

var (
	instance *Client
	once     sync.Once
)

// GetClient returns a singleton instance of the PostgreSQL client
func GetClient() (*Client, error) {
	var err error
	once.Do(func() {
		// Get configuration from environment variables
		host := getEnv("POSTGRES_HOST", "postgres")
		portStr := getEnv("POSTGRES_PORT", "5432")
		user := getEnv("POSTGRES_USER", "postgres")
		password := getEnv("POSTGRES_PASSWORD", "postgres123")
		dbName := getEnv("POSTGRES_DB", "tabular_data")
		sslMode := getEnv("POSTGRES_SSLMODE", "disable")

		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Printf("Invalid port number: %s, using default 5432", portStr)
			port = 5432
		}

		cfg := Config{
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
			DBName:   dbName,
			SSLMode:  sslMode,
		}

		instance, err = NewClient(cfg)
		if err != nil {
			log.Printf("Failed to initialize PostgreSQL client: %v", err)
		} else {
			log.Printf("PostgreSQL client initialized successfully")
		}
	})

	return instance, err
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// NewClient creates a new PostgreSQL client
func NewClient(cfg Config) (*Client, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	return NewClientFromDSN(dsn)
}

// NewClientFromConfig creates a new PostgreSQL client from a config struct
func NewClientFromConfig(cfg *config.PostgresConfig) (*Client, error) {
	port, err := strconv.Atoi(cfg.Port)
	if err != nil {
		port = 5432 // Default port
	}

	return NewClient(Config{
		Host:     cfg.Host,
		Port:     port,
		User:     cfg.User,
		Password: cfg.Password,
		DBName:   cfg.DBName,
		SSLMode:  cfg.SSLMode,
	})
}

// NewClientFromDSN creates a new PostgreSQL client from a connection string
func NewClientFromDSN(dsn string) (*Client, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	return &Client{db: db}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// DB returns the underlying *sql.DB instance
func (c *Client) DB() *sql.DB {
	return c.db
}
