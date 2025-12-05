package cmd

import (
	"os"
	"path"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/stashsphere/backend/config"
)

var openApiDumpCommand = &cobra.Command{
	Use:   "openapi-dump",
	Short: "Dump the OpenAPI schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath, _ := cmd.Flags().GetString("output")

		var config config.StashSphereServeConfig

		stateDir := os.Getenv("STATE_DIRECTORY")
		if stateDir == "" {
			stateDir = "."
		}
		cacheDir := os.Getenv("CACHE_DIRECTORY")
		if cacheDir == "" {
			cacheDir = "."
		}
		imagePath := path.Join(stateDir, "image_store")
		imageCachePath := path.Join(cacheDir, "image_cache")

		k := koanf.New(".")
		k.Load(confmap.Provider(map[string]interface{}{
			"database": map[string]interface{}{
				"user": "stashsphere",
				"name": "stashsphere",
				"host": "127.0.0.1",
			},
			"listenAddress": ":8081",
			"auth": map[string]interface{}{
				"privateKey": "",
			},
			"image": map[string]interface{}{
				"path":      imagePath,
				"cachePath": imageCachePath,
			},
			"invites": map[string]interface{}{
				"enabled": false,
				"code":    "",
			},
			"domains": map[string]interface{}{
				"allowed": []string{"http://localhost"},
				"own":     []string{"localhost"},
			},
			"frontendUrl":  "http://localhost",
			"instanceName": "stashsphereDev",
			"email": map[string]interface{}{
				"backend": "stdout",
			},
		}, "."), nil)

		if _, err := os.Stat("stashsphere.yaml"); err == nil {
			if err := k.Load(file.Provider("stashsphere.yaml"), yaml.Parser()); err != nil {
				log.Fatal().Msgf("error loading config: %v", err)
			}
		}
		k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: false})

		if outputPath == "" || outputPath == "-" {
			outputPath = "doc/openapi.json"
		}

		_, engine, db, err := setup(config, false, true, outputPath)
		if err != nil {
			return err
		}
		defer func() {
			if err := db.Close(); err != nil {
				log.Error().Err(err).Msg("failed to close database connection")
			}
		}()

		_ = engine.OutputOpenAPISpec()
		return nil
	},
}

func init() {
	openApiDumpCommand.Flags().StringP("output", "o", "", "output file path (default: stdout)")
	rootCmd.AddCommand(openApiDumpCommand)
}
