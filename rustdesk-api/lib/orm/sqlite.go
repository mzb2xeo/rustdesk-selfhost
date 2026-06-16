package orm

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type SqliteConfig struct {
	MaxIdleConns int
	MaxOpenConns int
}

func NewSqlite(sqliteConf *SqliteConfig, logwriter logger.Writer) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("./data/rustdeskapi.db?_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.New(
			logwriter, // io writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  logger.Warn, // Log level
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
				ParameterizedQueries:      true,        // Don't include params in the SQL log
				Colorful:                  true,
			},
		),
	})
	if err != nil {
		fmt.Println(err)
	}
	sqlDB, err2 := db.DB()
	if err2 != nil {
		fmt.Println(err2)
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
	sqlDB.SetMaxIdleConns(sqliteConf.MaxIdleConns)

	// SetMaxOpenConns sets the maximum number of open database connections.
	maxOpenConns := sqliteConf.MaxOpenConns
	if maxOpenConns <= 0 || maxOpenConns > 1 {
		maxOpenConns = 1
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)

	return db
}
