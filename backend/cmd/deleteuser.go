package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stashsphere/backend/operations"
)

var deleteUserCommand = &cobra.Command{
	Use:   "delete-user <email-or-id>",
	Short: "Schedule a user account for deletion",
	Long:  `Schedules the user identified by email or ID for immediate deletion by the purge worker.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		emailOrID := args[0]
		configPaths, _ := cmd.Flags().GetStringSlice("conf")
		yes, _ := cmd.Flags().GetBool("yes")
		minutes, _ := cmd.Flags().GetInt("minutes")

		db, err := openDatabase(configPaths)
		if err != nil {
			return err
		}
		defer db.Close()

		ctx := context.Background()

		var findErr error
		var userID, userName, userEmail string

		if strings.Contains(emailOrID, "@") {
			user, err := operations.FindUserByEmail(ctx, db, emailOrID)
			if err != nil {
				findErr = fmt.Errorf("user with email %q not found", emailOrID)
			} else {
				userID = user.ID
				userName = user.Name
				userEmail = user.Email
			}
		} else {
			user, err := operations.FindUserByID(ctx, db, emailOrID)
			if err != nil {
				findErr = fmt.Errorf("user with ID %q not found", emailOrID)
			} else {
				userID = user.ID
				userName = user.Name
				userEmail = user.Email
			}
		}
		if findErr != nil {
			return findErr
		}

		if !yes {
			fmt.Printf("Are you sure you want to delete user %q (%s)? [y/N] ", userName, userEmail)
			reader := bufio.NewReader(os.Stdin)
			answer, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading confirmation: %w", err)
			}
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		purgeAt := time.Now().UTC().Add(time.Duration(minutes) * time.Minute)

		_, err = operations.ScheduleUserDeletion(ctx, db, userID, purgeAt)
		if err != nil {
			return fmt.Errorf("error scheduling user deletion: %w", err)
		}

		fmt.Printf("User %q (%s) scheduled for deletion at %s\n", userName, userEmail, purgeAt.Format(time.RFC3339))

		return nil
	},
}

func init() {
	deleteUserCommand.Flags().StringSlice("conf", []string{"stashsphere.yaml"}, "path to one or more .yaml config files")
	deleteUserCommand.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
	deleteUserCommand.Flags().IntP("minutes", "m", 0, "delay deletion by this many minutes")
	rootCmd.AddCommand(deleteUserCommand)
}
