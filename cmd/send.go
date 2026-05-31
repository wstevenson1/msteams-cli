package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wstevenson1/msteams-cli/internal/auth"
	"github.com/wstevenson1/msteams-cli/internal/config"
	"github.com/wstevenson1/msteams-cli/internal/graph"
)

func newSendCmd() *cobra.Command {
	var to, message string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a Teams message to a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSend(cmd.Context(), to, message)
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "recipient email address (required)")
	cmd.Flags().StringVarP(&message, "message", "m", "", "message body (required)")
	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("message")

	return cmd
}

func runSend(ctx context.Context, to, message string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if cfg.ClientID == "" {
		return fmt.Errorf("no client ID configured; run with --client-id <id> on first use")
	}

	cred, err := auth.NewCredential(cfg.ClientID)
	if err != nil {
		return fmt.Errorf("creating credential: %w", err)
	}

	token, err := auth.GetToken(ctx, cred)
	if err != nil {
		return fmt.Errorf("authenticating: %w", err)
	}

	client := graph.New(token)

	me, err := client.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("getting current user: %w", err)
	}

	recipient, err := client.GetUserByEmail(ctx, to)
	if err != nil {
		return fmt.Errorf("user not found %q: %w", to, err)
	}

	chat, err := client.CreateOrGetChat(ctx, me.ID, recipient.ID)
	if err != nil {
		return fmt.Errorf("creating chat: %w", err)
	}

	if err := client.SendMessage(ctx, chat.ID, message); err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	fmt.Printf("Message sent to %s\n", to)
	return nil
}
