package database

import (
	"database/sql"

	"github.com/resonatecoop/id/config"
	bun "github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	bundebug "github.com/uptrace/bun/extra/bundebug"

	// Drivers
	_ "github.com/lib/pq"
	//_ "github.com/uptrace/bun/dialects/mysql"
)

func init() {
	// sql.NowFunc = func() time.Time {
	// 	return time.Now().UTC()
	// }
}

// NewDatabase returns a bun.DB struct
func NewDatabase(cnf *config.Config) (*bun.DB, error) {
	// Postgres
	sqldb, err := sql.Open("pgx", cnf.Database.PSN)

	if err != nil {
		panic(err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())

	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose()))

	if err != nil {
		return db, err
	}

	// Max idle connections
	// db.DB().SetMaxIdleConns(cnf.Database.MaxIdleConns)

	// // Max open connections
	// db.DB().SetMaxOpenConns(cnf.Database.MaxOpenConns)

	// // Database logging
	// db.LogMode(cnf.IsDevelopment)

	return db, nil
}
