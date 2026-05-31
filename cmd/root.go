package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wstevenson/msteams-cli/internal/config"
)

var clientID string

var rootCmd = &cobra.Command{
	Use:   "msteams-cli",
	Short: "Send Microsoft Teams messages from the command line",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if clientID != "" {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.ClientID = clientID
			return config.Save(cfg)
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&clientID, "client-id", "", "Azure app registration client ID (saved on first use)")
	rootCmd.AddCommand(newSendCmd())
}
