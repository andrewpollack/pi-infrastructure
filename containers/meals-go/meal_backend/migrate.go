package meal_backend

import (
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Postgres driver for migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"       // File source for migrate
)

// runMigrations applies the DB migrations from SQL files before starting the server.
func (c Config) runMigrations() {
	m, err := migrate.New(
		c.PostgresMigrationDir,
		fmt.Sprintf("%s?sslmode=disable", c.PostgresURL),
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrations: %v", err)
	}

	// Apply all up migrations. ErrNoChange is not considered an error.
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		// Check if the error is due to a dirty database
		if strings.Contains(err.Error(), "Dirty") {
			version, dirty, verErr := m.Version()
			if verErr != nil {
				log.Fatalf("Migration failed: cannot get version: %v", verErr)
			}
			if dirty {
				log.Printf("Database is dirty at version %d. Forcing migration version.", version)
				// Force the migration version to the current version to clear the dirty flag.
				if forceErr := m.Force(int(version)); forceErr != nil {
					log.Fatalf("Failed to force migration version: %v", forceErr)
				}
				// Retry the migration.
				if err = m.Up(); err != nil && err != migrate.ErrNoChange {
					log.Fatalf("Migration failed after forcing version: %v", err)
				}
			}
		} else {
			log.Fatalf("Migration failed: %v", err)
		}
	}

	log.Println("Migrations applied successfully!")
}
