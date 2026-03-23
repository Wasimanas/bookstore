package db

import (
	"database/sql"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"store/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)



func RunDBMigrations(db *sql.DB, sugar *zap.SugaredLogger) (error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		zap.L().Fatal("could not start sql migration", zap.Error(err))
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		zap.L().Fatal("migration failed", zap.Error(err))
        return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		zap.L().Fatal("An error occurred while syncing the database..", zap.Error(err))
        return err
	}

	zap.L().Info("Database migration ran successfully")
    return nil
}



func InitDB(cfg *config.ApplicationConfig) (*sql.DB, error) {

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		zap.L().Error("Failed to open DB connection", zap.Error(err))
		return nil, err
	}

	if err := db.Ping(); err != nil {
		zap.L().Error("Failed to ping DB", zap.Error(err))
		return nil, err
	}

	return db, nil
}
