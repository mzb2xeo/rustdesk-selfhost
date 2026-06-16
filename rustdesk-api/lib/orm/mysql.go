package orm

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type MysqlConfig struct {
	Dsn          string
	MaxIdleConns int
	MaxOpenConns int
}

func NewMysql(mysqlConf *MysqlConfig, logwriter logger.Writer) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:               mysqlConf.Dsn, // DSN data source name
		DefaultStringSize: 256,           // The default length of string type fields
		//DisableDatetimePrecision: true, // Disable datetime precision, which is not supported by databases before MySQL 5.6
		//DontSupportRenameIndex: true, // When renaming the index, delete and create a new one. Databases before MySQL 5.7 and MariaDB do not support renaming indexes.
		//DontSupportRenameColumn: true, // Use `change` to rename columns. Databases before MySQL 8 and MariaDB do not support renaming columns.
		//SkipInitializeWithVersion: false, // Automatically configure according to the current MySQL version
	}), &gorm.Config{
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
	sqlDB.SetMaxIdleConns(mysqlConf.MaxIdleConns)

	// SetMaxOpenConns sets the maximum number of open database connections.
	sqlDB.SetMaxOpenConns(mysqlConf.MaxOpenConns)

	return db
}
