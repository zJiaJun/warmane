package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func Open() (db *gorm.DB, err error) {
	db, err = gorm.Open(sqlite.Open("warmane.db"), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "[GORM]\u0020", log.Ldate|log.Lmicroseconds),
			logger.Config{
				SlowThreshold:             100 * time.Millisecond,
				Colorful:                  false,
				IgnoreRecordNotFoundError: false,
				ParameterizedQueries:      true,
				LogLevel:                  logger.Info,
			},
		),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		PrepareStmt:          true,
		DisableAutomaticPing: true,
		CreateBatchSize:      2000,
	})
	if err != nil {
		return
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetMaxOpenConns(100)
	}
	return
}
