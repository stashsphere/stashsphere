package cmd

import (
	"bufio"
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/stashsphere/backend/config"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
)

const passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
const generatedPasswordLength = 24

var resetPasswordCommand = &cobra.Command{
	Use:   "reset-password <email>",
	Short: "Reset a user's password",
	Long: `Reset a user's password by email address.

By default, prompts for a new password on STDIN.
Use -g to generate a secure random password instead.

Examples:
  # Prompt for password
  stashsphere reset-password user@example.com

  # Generate a random password
  stashsphere reset-password -g user@example.com

  # Pipe password from another command
  echo "newpassword123" | stashsphere reset-password user@example.com`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		email := args[0]
		configPaths, _ := cmd.Flags().GetStringSlice("conf")
		generate, _ := cmd.Flags().GetBool("generate")

		db, err := openDatabase(configPaths)
		if err != nil {
			return err
		}
		defer db.Close()

		ctx := context.Background()

		user, err := models.Users(models.UserWhere.Email.EQ(email)).One(ctx, db)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("user with email %q not found", email)
			}
			return fmt.Errorf("error finding user: %w", err)
		}

		var password string
		if generate {
			password, err = generateSecurePassword(generatedPasswordLength)
			if err != nil {
				return fmt.Errorf("error generating password: %w", err)
			}
		} else {
			password, err = readPasswordFromStdin()
			if err != nil {
				return fmt.Errorf("error reading password: %w", err)
			}
			if password == "" {
				return fmt.Errorf("no password provided")
			}
		}

		passwordHash, err := operations.HashPassword(password)
		if err != nil {
			return fmt.Errorf("error hashing password: %w", err)
		}

		user.PasswordHash = string(passwordHash)
		_, err = user.Update(ctx, db, boil.Whitelist(models.UserColumns.PasswordHash))
		if err != nil {
			return fmt.Errorf("error updating password: %w", err)
		}

		if generate {
			fmt.Println(password)
		} else {
			log.Info().Str("email", email).Msg("password reset successfully")
		}

		return nil
	},
}

func openDatabase(configPaths []string) (*sql.DB, error) {
	var conf config.StashSphereMigrateConfig

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
		k.UnmarshalWithConf("", &conf, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: false})
	}

	dbOptions := fmt.Sprintf("user=%s dbname=%s host=%s", conf.Database.User, conf.Database.Name, conf.Database.Host)
	if conf.Database.Password != nil {
		dbOptions = fmt.Sprintf("%s password=%s", dbOptions, *conf.Database.Password)
	}
	if conf.Database.Port != nil {
		dbOptions = fmt.Sprintf("%s port=%d", dbOptions, *conf.Database.Port)
	}
	if conf.Database.SslMode != nil {
		dbOptions = fmt.Sprintf("%s sslmode=%s", dbOptions, *conf.Database.SslMode)
	}

	return sql.Open("postgres", dbOptions)
}

func readPasswordFromStdin() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	// Check if there's data on stdin (piped or redirected)
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return strings.TrimSpace(scanner.Text()), nil
		}
		if err := scanner.Err(); err != nil {
			return "", err
		}
	}

	return "", nil
}

func generateSecurePassword(length int) (string, error) {
	password := make([]byte, length)
	charsetLen := big.NewInt(int64(len(passwordCharset)))

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		password[i] = passwordCharset[idx.Int64()]
	}

	return string(password), nil
}

func init() {
	resetPasswordCommand.Flags().StringSlice("conf", []string{"stashsphere.yaml"}, "path to one or more .yaml config files")
	resetPasswordCommand.Flags().BoolP("generate", "g", false, "generate a secure random password")
	rootCmd.AddCommand(resetPasswordCommand)
}
