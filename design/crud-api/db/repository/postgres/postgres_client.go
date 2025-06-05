package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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

// NewClient creates a new PostgreSQL client
func NewClient(cfg Config) (*Client, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	return NewClientFromDSN(dsn)
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

// InitializeTables creates the necessary tables if they don't exist
func (c *Client) InitializeTables(ctx context.Context) error {
	// Create entity_attributes table
	entityAttributesSQL := `
	CREATE TABLE IF NOT EXISTS entity_attributes (
		id SERIAL PRIMARY KEY,
		entity_id VARCHAR(255) NOT NULL,
		attribute_name VARCHAR(255) NOT NULL,
		table_name VARCHAR(255) NOT NULL,
		schema_version INT NOT NULL DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(entity_id, attribute_name)
	);`

	// Create attribute_schemas table
	attributeSchemasSQL := `
	CREATE TABLE IF NOT EXISTS attribute_schemas (
		id SERIAL PRIMARY KEY,
		table_name VARCHAR(255) NOT NULL,
		schema_version INT NOT NULL,
		schema_definition JSONB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(table_name, schema_version)
	);`

	// Execute the creation queries
	if _, err := c.db.ExecContext(ctx, entityAttributesSQL); err != nil {
		return fmt.Errorf("error creating entity_attributes table: %v", err)
	}

	if _, err := c.db.ExecContext(ctx, attributeSchemasSQL); err != nil {
		return fmt.Errorf("error creating attribute_schemas table: %v", err)
	}

	return nil
}

// TableExists checks if a table exists in the database
func (c *Client) TableExists(ctx context.Context, tableName string) (bool, error) {
	query := `
	SELECT EXISTS (
		SELECT FROM pg_tables
		WHERE schemaname = 'public'
		AND tablename = $1
	);`

	var exists bool
	err := c.db.QueryRowContext(ctx, query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking table existence: %v", err)
	}

	return exists, nil
}

// CreateDynamicTable creates a new table for storing attribute data
func (c *Client) CreateDynamicTable(ctx context.Context, tableName string, columns []Column) error {
	// Build column definitions
	var columnDefs []string
	
	// Add primary key and entity_attribute_id first
	columnDefs = append(columnDefs, "id SERIAL PRIMARY KEY")
	columnDefs = append(columnDefs, "entity_attribute_id INTEGER REFERENCES entity_attributes(id)")
	
	// Add the rest of the columns
	for _, col := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", col.Name, col.Type))
	}
	
	// Add created_at timestamp at the end
	columnDefs = append(columnDefs, "created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP")

	// Create table query
	createTableSQL := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		%s
	);`, tableName, strings.Join(columnDefs, ",\n"))

	// Execute the creation query
	if _, err := c.db.ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("error creating dynamic table: %v", err)
	}

	return nil
}

// InsertTabularData inserts rows into a dynamic table
func (c *Client) InsertTabularData(ctx context.Context, tableName string, entityAttributeID int, columns []string, rows [][]interface{}) error {
	// Build the INSERT query
	columnNames := append([]string{"entity_attribute_id"}, columns...)
	placeholders := make([]string, len(rows))
	valuesPerRow := len(columns) + 1 // +1 for entity_attribute_id

	for i := range rows {
		rowPlaceholders := make([]string, valuesPerRow)
		for j := range rowPlaceholders {
			rowPlaceholders[j] = fmt.Sprintf("$%d", i*valuesPerRow+j+1)
		}
		placeholders[i] = fmt.Sprintf("(%s)", strings.Join(rowPlaceholders, ", "))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(columnNames, ", "),
		strings.Join(placeholders, ", "),
	)

	// Flatten values for the query
	values := make([]interface{}, 0, len(rows)*valuesPerRow)
	for _, row := range rows {
		values = append(values, entityAttributeID) // Add entity_attribute_id first
		values = append(values, row...)
	}

	// Execute the query
	_, err := c.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("error inserting data: %v", err)
	}

	return nil
}
