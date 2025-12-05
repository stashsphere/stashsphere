package testcommon

import (
	"fmt"
	"log"
	"os"

	"database/sql"

	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func CreateTestSchema() (*sql.DB, func(), error) {
	schema_id, err := gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", 32)
	if err != nil {
		return nil, nil, err
	}
	schemaName := fmt.Sprintf("test_%s", schema_id)
	log.Printf("Running in %s", schemaName)

	pgHost := os.Getenv("PGHOST")
	pgPort := os.Getenv("PGPORT")
	pgUser := os.Getenv("PGUSER")
	pgDatabase := os.Getenv("PGDATABASE")
	pgPassword := os.Getenv("PGPASSWORD")
	if pgPassword != "" {
		pgPassword = "&password=" + pgPassword
	}
	if pgPort != "" {
		pgPort = "&port=" + pgPort
	}

	basePath := fmt.Sprintf("postgres:///%s?host=%s%s&user=%s%s&sslmode=disable", pgDatabase, pgHost, pgPort, pgUser, pgPassword)

	log.Printf("Database basePath: %s", basePath)

	rootDb, err := sql.Open("postgres", basePath)
	if err != nil {
		return nil, nil, err
	}
	_, err = rootDb.Exec(fmt.Sprintf("CREATE SCHEMA %s", schemaName))
	if err != nil {
		return nil, nil, err
	}
	teardownFunc := func() {
		_, err = rootDb.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE", schemaName))
		if err != nil {
			log.Fatal(err)
		}
		err = rootDb.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
	final_db_path := fmt.Sprintf("%s&search_path=%s", basePath, schemaName)
	db, err := sql.Open("postgres", final_db_path)
	if err != nil {
		return nil, nil, err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://../migrations",
		"postgres", driver)
	if err != nil {
		return nil, nil, err
	}
	m.Up()
	return db, teardownFunc, nil
}
