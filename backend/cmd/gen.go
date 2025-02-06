package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stashsphere/backend/crypto"
)

var genCommand = &cobra.Command{
	Use:   "genkey",
	Short: "generate a new ed25519 key",
	RunE: func(cmd *cobra.Command, args []string) error {
		privateKey, err := crypto.GenerateEd25519StringKey()
		if err != nil {
			return err
		}
		fmt.Printf("Generated Private Key: %s\n", privateKey)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genCommand)
}
