package testcommon

import (
	"fmt"
	"log"

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
	schema_name := fmt.Sprintf("test_%s", schema_id)
	log.Printf("Running in %s", schema_name)
	root_db, err := sql.Open("postgres", "postgres://stashsphere:secret@localhost:5432/stashsphere?sslmode=disable")
	if err != nil {
		return nil, nil, err
	}
	_, err = root_db.Exec(fmt.Sprintf("CREATE SCHEMA %s", schema_name))
	if err != nil {
		return nil, nil, err
	}
	teardownFunc := func() {
		_, err = root_db.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE", schema_name))
		if err != nil {
			log.Fatal(err)
		}
	}
	final_db_path := fmt.Sprintf("postgres://stashsphere:secret@localhost:5432/stashsphere?sslmode=disable&search_path=%s", schema_name)
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
