package db

import (
	"database/sql"

	"github.com/CPU-commits/Template_Go-EventDriven/src/package/logger"
	"github.com/CPU-commits/Template_Go-EventDriven/src/settings"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var settingsData = settings.GetSettings()

func Init(logger logger.Logger) {
	db, err := sql.Open("pgx", settingsData.DB_CONNECTION)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}

	boil.SetDB(db)
	logger.Info("Database Connected")
}
