package database

import (
	"fmt"
	"time"

	"github.com/RichardKnop/go-oauth2-server/config"
	"github.com/jinzhu/gorm"

	// Drivers
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/lib/pq"
)

func init() {
	gorm.NowFunc = func() time.Time {
		return time.Now().UTC()
	}
}

// NewDatabase returns a gorm.DB struct, gorm.DB.DB() returns a database handle
// see http://golang.org/pkg/database/sql/#DB
func NewDatabase(cnf *config.Config) (*gorm.DB, error) {
	// Postgres
	if cnf.Database.Type == "postgres" {
		// Connection args
		// see https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
		args := fmt.Sprintf(
			"sslmode=disable host=%s port=%d user=%s password='%s' dbname=%s",
			cnf.Database.Host,
			cnf.Database.Port,
			cnf.Database.User,
			cnf.Database.Password,
			cnf.Database.DatabaseName,
		)

		db, err := gorm.Open(cnf.Database.Type, args)
		if err != nil {
			return db, err
		}

		// Max idle connections
		db.DB().SetMaxIdleConns(cnf.Database.MaxIdleConns)

		// Max open connections
		db.DB().SetMaxOpenConns(cnf.Database.MaxOpenConns)

		// Database logging
		db.LogMode(cnf.IsDevelopment)

		return db, nil
	}

	// Database type not supported
	return nil, fmt.Errorf("Database type %s not suppported", cnf.Database.Type)
}

// NewDatabase2 returns a gorm.DB struct, gorm.DB.DB() returns a database handle
// Alternate database config for mysql wordpress database and/or future posgres user-api db
// see http://golang.org/pkg/database/sql/#DB
func NewDatabase2(cnf *config.Config) (*gorm.DB, error) {
	if cnf.Database2.Type == "postgres" {
		// Connection args
		// see https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
		args := fmt.Sprintf(
			"sslmode=disable host=%s port=%d user=%s password='%s' dbname=%s",
			cnf.Database2.Host,
			cnf.Database2.Port,
			cnf.Database2.User,
			cnf.Database2.Password,
			cnf.Database2.DatabaseName,
		)

		db, err := gorm.Open(cnf.Database2.Type, args)
		if err != nil {
			return db, err
		}

		// Max idle connections
		db.DB().SetMaxIdleConns(cnf.Database2.MaxIdleConns)

		// Max open connections
		db.DB().SetMaxOpenConns(cnf.Database2.MaxOpenConns)

		// Database logging
		db.LogMode(cnf.IsDevelopment)

		return db, nil
	}

	if cnf.Database2.Type == "mysql" {
		args := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True",
			cnf.Database2.User,
			cnf.Database2.Password,
			cnf.Database2.Host,
			cnf.Database2.Port,
			cnf.Database2.DatabaseName,
		)
		db, err := gorm.Open("mysql", args)
		if err != nil {
			return db, err
		}
		// Database logging
		db.LogMode(cnf.IsDevelopment)
		return db, nil
	}

	// Database type not supported
	return nil, fmt.Errorf("Secondary database type %s not suppported", cnf.Database2.Type)
}
