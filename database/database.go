package database

import (
	"github.com/jackc/pgx/v4"
	stdlib "github.com/jackc/pgx/v4/stdlib"
	"github.com/resonatecoop/id/config"
	bun "github.com/uptrace/bun"

	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"

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
	dbconfig, err := pgx.ParseConfig(cnf.Database.PSN)

	if err != nil {
		return nil, err
	}

	dbconfig.PreferSimpleProtocol = true

	sqldb := stdlib.OpenDB(*dbconfig)

	db := bun.NewDB(sqldb, pgdialect.New())

	if cnf.IsDevelopment {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	_, err = db.Exec("SELECT 1=1")

	if err != nil {
		return db, err
	}

	return db, err
}
