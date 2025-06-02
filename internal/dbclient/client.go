package dbclient

import (
	"context"
	"embedup-go/configs/config"
	"embedup-go/internal/cstmerr"
	"fmt"
	"time"
)

// QueryOptions can be used to pass common query modifiers like limit, offset, order.
// This is a simplified example; a real-world scenario might use a more structured approach
// or ORM-specific option builders if full abstraction is difficult.
type QueryOptions struct {
	Limit  int
	Offset int
	Order  string // e.g., "created_at desc"
	// Preloads []string // For eager loading relationships, e.g., Preloads: []string{"UserProfile", "Orders"}
	// SelectFields []string // To select specific fields
}

// DBClient defines the interface for ORM-like database operations.
type DBClient interface {
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	// Create inserts a new record into the database.
	// 'model' is a pointer to the struct to be created.
	Create(ctx context.Context, model interface{}) error

	// Save updates an existing record or creates it if it does not exist (upsert-like or based on primary key).
	// Behavior can be ORM-dependent. GORM's Save updates if PK is set, otherwise creates.
	// 'model' is a pointer to the struct to be saved.
	Save(ctx context.Context, model interface{}) error

	// Updates updates attributes for a record.
	// 'model' is a pointer to the struct (can be a partial struct or map for updates).
	// 'conditionModel' is optional, a pointer to a struct with PK or unique fields to identify the record to update.
	// If 'conditionModel' is nil, 'model' itself must contain the PK.
	// 'data' can be a struct or map[string]interface{} for the fields to update.
	Updates(ctx context.Context, modelWithPK interface{}, data interface{}) error

	// Delete deletes a record.
	// 'model' is a pointer to the struct with its primary key set, or a struct defining conditions.
	Delete(ctx context.Context, model interface{}, conditions ...interface{}) error // conditions can be id, or query + args

	// First retrieves the first record matching the given conditions.
	// 'model' is a pointer to the struct to scan data into.
	// 'conditions' can be a primary key, a struct to build WHERE conditions, or query string + args.
	// The interpretation of 'conditions' will be up to the adapter.
	First(ctx context.Context, model interface{}, conditions ...interface{}) error

	// Find retrieves a collection of models matching the given conditions.
	// 'collection' is a pointer to a slice of structs.
	// 'conditions' can be a struct to build WHERE conditions, or query string + args.
	Find(ctx context.Context, collection interface{}, conditions ...interface{}) error

	// ExecRaw executes a raw SQL query that doesn't necessarily map directly to a model.
	// Kept for flexibility (e.g., complex joins, DDL, functions not covered by ORM methods).
	ExecRaw(ctx context.Context, query string, args ...interface{}) (QueryResult, error)

	// SelectRaw executes a raw SQL query and scans results into models.
	// Kept for flexibility.
	SelectRaw(ctx context.Context, collectionOrModel interface{}, query string, args ...interface{}) error

	// RunInTransaction executes a function within a database transaction.
	RunInTransaction(ctx context.Context, fn func(ctx context.Context, txClient DBClient) error) error

	// TODO: Consider adding methods for:
	// - Count(ctx context.Context, model interface{}, conditions ...interface{}) (int64, error)
	// - Pluck(ctx context.Context, model interface{}, fieldName string, dest interface{}, conditions ...interface{}) error
	// - Advanced querying with QueryOptions or a builder pattern.
}

// QueryResult (can remain the same for ExecRaw)
type QueryResult interface {
	RowsAffected() int
}

// NewDBClient is a factory function that will return a specific DBClient implementation.
func NewDBClient(dbConfig *config.DatabaseConfig, dbType string) (DBClient, error) { // Added dbType
	if dbConfig == nil {
		return nil, fmt.Errorf("database configuration is nil")
	}

	var adapter DBClient
	var err error

	// Example: Use a type string from config or argument to switch adapters
	// For now, we'll explicitly choose GORM.
	// You could add a field like `dbConfig.Type` ("gorm", "pg", etc.)
	switch dbType {
	case "gorm":
		gormAdapter := NewGORMAdapter(dbConfig) // Use the new GORM adapter
		adapter = gormAdapter
	// case "pg": // Keep previous pg_adapter logic if needed, or remove if GORM is the sole focus now
	// 	pgAdapter := NewPGAdapter(dbConfig)
	// 	adapter = pgAdapter
	default:
		return nil, cstmerr.NewDBConnectionError(fmt.Sprintf("failed to find db type %s", dbType), err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Connection timeout
	defer cancel()

	if err = adapter.Connect(ctx); err != nil {
		// The specific adapter's Connect method will wrap errors appropriately.
		return nil, cstmerr.NewDBConnectionError("failed to connect with pg_adapter", err)
	}
	return adapter, nil
}
