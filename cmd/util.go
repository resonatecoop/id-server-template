package cmd

import (
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/database"

	//	"github.com/resonatecoop/id/database"
	"github.com/uptrace/bun"
)

// initConfigDB loads the configuration and connects to the database
func initConfigDB(mustLoadOnce, keepReloading bool, configBackend string) (*config.Config, *bun.DB, error) {
	// Config
	cnf := config.NewConfig(mustLoadOnce, keepReloading, configBackend)

	// Database
	db, err := database.NewDatabase(cnf)
	if err != nil {
		return nil, nil, err
	}

	return cnf, db, nil
}
