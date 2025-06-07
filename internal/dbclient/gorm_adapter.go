// internal/dbclient/gorm_adapter.go
package dbclient

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"

	"embedup-go/configs/config"
	"embedup-go/internal/cstmerr"
	"embedup-go/internal/shared"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func pascalToCamelCase(s string) string {
	if len(s) == 0 {
		return ""
	}

	words := make([]string, 0)
	currentWord := strings.Builder{}
	for i, r := range s {
		if unicode.IsUpper(r) {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
		currentWord.WriteRune(r)
		if i == len(s)-1 {
			words = append(words, currentWord.String())
		}
	}

	if len(words) == 0 {
		return ""
	}

	words[0] = strings.ToLower(words[0])
	return strings.Join(words, "")
}

// GORMAdapter implements the DBClient interface using the GORM library.
type GORMAdapter struct {
	db     *gorm.DB
	config *config.DatabaseConfig
}

type CustomNamingStrategy struct {
	schema.NamingStrategy // Embed the default naming strategy
}

func (c CustomNamingStrategy) ColumnName(table, column string) string {
	//convert column to camelCase
	return pascalToCamelCase(column)
}

// NewGORMAdapter creates a new GORMAdapter.
func NewGORMAdapter(cfg *config.DatabaseConfig) *GORMAdapter {
	return &GORMAdapter{
		config: cfg,
	}
}

// Connect, Close, Ping methods remain the same as in the previous GORM adapter.
func (ga *GORMAdapter) Connect(ctx context.Context) error {
	if ga.db != nil {
		sqlDB, err := ga.db.DB()
		if err == nil {
			if err = sqlDB.PingContext(ctx); err == nil {
				return nil
			}
		}
	}
	// 	NOTE: The following commented code is an example of how to create a database if it doesn't exist.
	createDBDsn := fmt.Sprintf("host=%s user=%s password=%s port=%d sslmode=%s TimeZone=UTC",
		ga.config.Host, ga.config.User, ga.config.Password, // Ensure this is PasswordConf
		ga.config.Port, ga.config.SSLMode)

	database, _ := gorm.Open(postgres.Open(createDBDsn), &gorm.Config{})
	_ = database.Exec("CREATE DATABASE " + ga.config.DBName + ";")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		ga.config.Host, ga.config.User, ga.config.Password, // Ensure this is PasswordConf
		ga.config.DBName, ga.config.Port, ga.config.SSLMode)

	gormLogger := logger.New(log.New(log.Writer(), "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold: time.Second, LogLevel: logger.Warn, IgnoreRecordNotFoundError: true, Colorful: false,
	})

	var err error
	ga.db, err = gorm.Open(postgres.Open(dsn),
		&gorm.Config{Logger: gormLogger,
			NowFunc: func() time.Time { return time.Now().UTC() },
			NamingStrategy: CustomNamingStrategy{
				schema.NamingStrategy{
					SingularTable: true,
				}}})
	if err != nil {
		return cstmerr.NewDBConnectionError("gorm.Open failed", err)
	}

	// TODO: Uncomment if you want to auto-migrate models
	ga.db.AutoMigrate(&shared.Updater{})
	// ga.db.AutoMigrate(shared.AutoMigrateList...)
	// err = ga.db.SetupJoinTable(&shared.Page{}, "Tabs", &shared.PageTabsTab{})
	// if err != nil {
	// 	return cstmerr.NewDBConnectionError("failed to setup join table for Page and Tabs", err)
	// }

	sqlDB, err := ga.db.DB()
	if err != nil {
		return cstmerr.NewDBConnectionError("failed to get underlying sql.DB from GORM", err)
	}
	if err = sqlDB.PingContext(ctx); err != nil {
		return cstmerr.NewDBConnectionError("failed to ping database after GORM connect", err)
	}
	fmt.Println("Successfully connected to PostgreSQL using GORM!")
	return nil
}

func (ga *GORMAdapter) Close() error {
	if ga.db != nil {
		sqlDB, _ := ga.db.DB()
		if sqlDB != nil {
			return sqlDB.Close()
		}
	}
	return nil
}

func (ga *GORMAdapter) Ping(ctx context.Context) error {
	if ga.db == nil {
		return cstmerr.NewDBError("database not connected (GORM)", nil)
	}
	sqlDB, _ := ga.db.DB()
	if sqlDB == nil {
		return cstmerr.NewDBError("underlying sql.DB not available for ping (GORM)", nil)
	}
	return sqlDB.PingContext(ctx)
}

// --- ORM-like methods ---

func (ga *GORMAdapter) Create(ctx context.Context, model interface{}) error {
	if ga.db == nil {
		return cstmerr.NewDBError("database not connected (GORM)", nil)
	}
	result := ga.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return cstmerr.NewDBQueryError("GORM Create failed", result.Error)
	}
	return nil
}

func (ga *GORMAdapter) Save(ctx context.Context, model interface{}) error {
	if ga.db == nil {
		return cstmerr.NewDBError("database not connected (GORM)", nil)
	}
	result := ga.db.WithContext(ctx).Save(model)
	if result.Error != nil {
		return cstmerr.NewDBQueryError("GORM Save failed", result.Error)
	}
	return nil
}

// Updates updates attributes for a record.
// 'modelWithPK' identifies the record (e.g. User{ID: 1})
// 'data' is a struct or map for the fields to update (e.g. User{Name: "new name"}, or map[string]interface{}{"name": "new name"})
func (ga *GORMAdapter) Updates(ctx context.Context, modelWithPK interface{}, data interface{}) error {
	if ga.db == nil {
		return cstmerr.NewDBError("database not connected (GORM)", nil)
	}
	// GORM's Updates method requires the model to infer the table.
	// The 'modelWithPK' helps scope the update if it contains the primary key.
	// If modelWithPK is just an ID, you might need Model(&SomeModelType{}).Where("id = ?", id).Updates(data)
	// For simplicity, this assumes modelWithPK is a struct that GORM can use to find the record by PK.
	result := ga.db.WithContext(ctx).Model(modelWithPK).Updates(data)
	if result.Error != nil {
		return cstmerr.NewDBQueryError("GORM Updates failed", result.Error)
	}
	if result.RowsAffected == 0 {
		// This might not be an error if "no update needed" is valid,
		// but could also mean the record to update wasn't found.
		// GORM doesn't treat 0 RowsAffected on Update as ErrRecordNotFound by default.
		// Consider if you need to explicitly check for record existence first or if this behavior is acceptable.
		// For now, we don't return an error here, but you might want to.
		// return cstmerr.NewDBNotFoundError("GORM Updates affected 0 rows, record possibly not found or no change", nil)
	}
	return nil
}

func (ga *GORMAdapter) Delete(ctx context.Context, model interface{}, conditions ...interface{}) error {
	if ga.db == nil {
		return cstmerr.NewDBError("database not connected (GORM)", nil)
	}
	// GORM's Delete:
	// db.Delete(&User{ID: 10})
	// db.Delete(&User{}, 10)
	// db.Delete(&User{}, "email = ?", "jinzhu@example.org")
	// db.Delete(&User{}, []int{1,2,3})
	// The 'model' argument provides the type (for table name) and potentially the PK.
	// 'conditions' are additional query conditions.
	var result *gorm.DB
	if len(conditions) > 0 {
		result = ga.db.WithContext(ctx).Delete(model, conditions...)
	} else {
		// If no conditions, GORM deletes based on primary key in 'model'
		// or deletes all records if model is an empty struct (dangerous, usually add a Where clause).
		// This assumes 'model' itself contains the primary key for deletion.
		result = ga.db.WithContext(ctx).Delete(model)
	}

	if result.Error != nil {
		return cstmerr.NewDBQueryError("GORM Delete failed", result.Error)
	}
	// Optionally check result.RowsAffected if you need to confirm something was deleted.
	// if result.RowsAffected == 0 {
	// 	return cstmerr.NewDBNotFoundError("GORM Delete affected 0 rows, record possibly not found", nil)
	// }
	return nil
}

func (ga *GORMAdapter) First(ctx context.Context, model interface{}, conditions ...interface{}) error {
	if ga.db == nil {
		return cstmerr.NewDBError("database not connected (GORM)", nil)
	}
	db := ga.db.WithContext(ctx)
	var result *gorm.DB
	if len(conditions) > 0 {
		result = db.First(model, conditions...)
	} else {
		// If no conditions, GORM might fetch the first record by primary key order.
		// Usually, First is called with conditions.
		result = db.First(model)
	}

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return cstmerr.NewDBNotFoundError("GORM First failed, record not found", result.Error)
		}
		return cstmerr.NewDBQueryError("GORM First failed", result.Error)
	}
	return nil
}

func (ga *GORMAdapter) Find(ctx context.Context, collection interface{}, conditions ...interface{}) error {
	if ga.db == nil {
		return cstmerr.NewDBError("database not connected (GORM)", nil)
	}
	// GORM's Find:
	// db.Find(&users, "name <> ?", "jinzhu")
	// db.Find(&users, User{Role: "admin"})
	db := ga.db.WithContext(ctx)
	var result *gorm.DB
	if len(conditions) > 0 {
		result = db.Find(collection, conditions...)
	} else {
		result = db.Find(collection) // Find all records for the given model type
	}

	if result.Error != nil {
		// GORM's Find doesn't typically return ErrRecordNotFound for an empty result set,
		// the slice/collection will just be empty.
		return cstmerr.NewDBQueryError("GORM Find failed", result.Error)
	}
	return nil
}

// --- Raw SQL methods ---
type gormQueryResult struct { // Re-define if not already in this file from previous version
	rowsAffected int64
}

func (r *gormQueryResult) RowsAffected() int {
	return int(r.rowsAffected)
}

func (ga *GORMAdapter) ExecRaw(ctx context.Context, query string, args ...interface{}) (QueryResult, error) {
	if ga.db == nil {
		return nil, cstmerr.NewDBError("database not connected (GORM)", nil)
	}
	result := ga.db.WithContext(ctx).Exec(query, args...)
	if result.Error != nil {
		return nil, cstmerr.NewDBQueryError(fmt.Sprintf("GORM ExecRaw query failed: %s", query), result.Error)
	}
	return &gormQueryResult{rowsAffected: result.RowsAffected}, nil
}

func (ga *GORMAdapter) SelectRaw(ctx context.Context, collectionOrModel interface{}, query string, args ...interface{}) error {
	if ga.db == nil {
		return cstmerr.NewDBError("database not connected (GORM)", nil)
	}
	result := ga.db.WithContext(ctx).Raw(query, args...).Scan(collectionOrModel)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound { // Raw can also return this if Scan expects one row
			return cstmerr.NewDBNotFoundError(fmt.Sprintf("GORM SelectRaw query (Scan) found no records: %s", query), result.Error)
		}
		return cstmerr.NewDBQueryError(fmt.Sprintf("GORM SelectRaw query failed: %s", query), result.Error)
	}
	// For Raw().Scan(&slice), if no rows, slice is empty, no error.
	// If Raw().Scan(&struct), and no rows, result.Error will be gorm.ErrRecordNotFound if RowsAffected is 0
	if result.RowsAffected == 0 && result.Error == nil {
		// If scanning into a single struct and no rows are found, GORM might set result.Error to gorm.ErrRecordNotFound.
		// If scanning into a slice, it will be empty.
		// We check if the error is already set; if not, and scanning into a single struct (not a slice),
		// and 0 rows affected, it implies not found. This logic is tricky to make generic here.
		// The current check `result.Error == gorm.ErrRecordNotFound` should cover most cases for single struct scan.
	}
	return nil
}

// --- Transaction method ---
// gormTxAdapter and RunInTransaction remain structurally similar to the previous GORM adapter
// but will now call the ORM-like methods of the gormTxAdapter.

type gormTxAdapter struct {
	tx *gorm.DB
}

func (gta *gormTxAdapter) Connect(ctx context.Context) error { /* ... */
	return cstmerr.NewDBError("cannot connect in tx", nil)
}
func (gta *gormTxAdapter) Close() error { /* ... */
	return cstmerr.NewDBError("cannot close in tx", nil)
}
func (gta *gormTxAdapter) Ping(ctx context.Context) error { /* ... */ return nil }

func (gta *gormTxAdapter) Create(ctx context.Context, model interface{}) error {
	return gta.tx.WithContext(ctx).Create(model).Error
}
func (gta *gormTxAdapter) Save(ctx context.Context, model interface{}) error {
	return gta.tx.WithContext(ctx).Save(model).Error
}
func (gta *gormTxAdapter) Updates(ctx context.Context, modelWithPK interface{}, data interface{}) error {
	return gta.tx.WithContext(ctx).Model(modelWithPK).Updates(data).Error
}
func (gta *gormTxAdapter) Delete(ctx context.Context, model interface{}, conditions ...interface{}) error {
	if len(conditions) > 0 {
		return gta.tx.WithContext(ctx).Delete(model, conditions...).Error
	}
	return gta.tx.WithContext(ctx).Delete(model).Error
}
func (gta *gormTxAdapter) First(ctx context.Context, model interface{}, conditions ...interface{}) error {
	var result *gorm.DB
	if len(conditions) > 0 {
		result = gta.tx.WithContext(ctx).First(model, conditions...)
	} else {
		result = gta.tx.WithContext(ctx).First(model)
	}
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		return cstmerr.NewDBNotFoundError("GORM First (TX) not found", result.Error)
	}
	return result.Error
}
func (gta *gormTxAdapter) Find(ctx context.Context, collection interface{}, conditions ...interface{}) error {
	var result *gorm.DB
	if len(conditions) > 0 {
		result = gta.tx.WithContext(ctx).Find(collection, conditions...)
	} else {
		result = gta.tx.WithContext(ctx).Find(collection)
	}
	return result.Error
}
func (gta *gormTxAdapter) ExecRaw(ctx context.Context, query string, args ...interface{}) (QueryResult, error) {
	res := gta.tx.WithContext(ctx).Exec(query, args...)
	if res.Error != nil {
		return nil, res.Error
	}
	return &gormQueryResult{rowsAffected: res.RowsAffected}, nil
}
func (gta *gormTxAdapter) SelectRaw(ctx context.Context, collectionOrModel interface{}, query string, args ...interface{}) error {
	res := gta.tx.WithContext(ctx).Raw(query, args...).Scan(collectionOrModel)
	if res.Error != nil && res.Error == gorm.ErrRecordNotFound {
		return cstmerr.NewDBNotFoundError("GORM SelectRaw (TX) not found", res.Error)
	}
	return res.Error
}
func (gta *gormTxAdapter) RunInTransaction(ctx context.Context, fn func(ctx context.Context, txClient DBClient) error) error {
	return cstmerr.NewDBError("nested transactions not directly supported by this basic GORM tx adapter", nil)
}

func (ga *GORMAdapter) RunInTransaction(ctx context.Context, fn func(ctx context.Context, txClient DBClient) error) error {
	if ga.db == nil {
		return cstmerr.NewDBError("database not connected (GORM)", nil)
	}
	return ga.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txAdapter := &gormTxAdapter{tx: tx}
		return fn(ctx, txAdapter)
	})
}
