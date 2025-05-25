package cmd

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/stashsphere/backend/config"
	"github.com/stashsphere/backend/migrations"
)

var migrateCommand = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the Postgresql Database",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPaths, _ := cmd.Flags().GetStringSlice("conf")

		var config config.StashSphereMigrateConfig

		k := koanf.New(".")
		k.Load(confmap.Provider(map[string]interface{}{
			"database": map[string]interface{}{
				"user": "stashsphere",
				"name": "stashsphere",
				"host": "127.0.0.1",
			},
		}, "."), nil)

		for _, configPath := range configPaths {
			if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
				log.Fatal().Msgf("error loading config: %v", err)
			}
			k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: false})
		}
		dbOptions := fmt.Sprintf("user=%s dbname=%s host=%s", config.Database.User, config.Database.Name, config.Database.Host)
		if config.Database.Password != nil {
			dbOptions = fmt.Sprintf("%s password=%s", dbOptions, *config.Database.Password)
		}
		if config.Database.Port != nil {
			dbOptions = fmt.Sprintf("%s port=%d", dbOptions, *config.Database.Port)
		}
		if config.Database.SslMode != nil {
			dbOptions = fmt.Sprintf("%s sslmode=%s", dbOptions, *config.Database.SslMode)
		}

		db, err := sql.Open("postgres", dbOptions)
		if err != nil {
			return err
		}
		migrationDir, err := iofs.New(migrations.FS, ".")
		if err != nil {
			return err
		}

		driver, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			return err
		}
		m, err := migrate.NewWithInstance("iofs", migrationDir, config.Database.Name, driver)
		if err != nil {
			return err
		}
		err = m.Up()
		if err := m.Up(); errors.Is(err, migrate.ErrNoChange) {
			log.Warn().Msg("no changes to apply, schema left unchanged")
			return nil
		}
		return err
	},
}

func init() {
	migrateCommand.Flags().StringSlice("conf", []string{"stashsphere.yaml"}, "path to one or more .yaml config files")
	rootCmd.AddCommand(migrateCommand)
}
