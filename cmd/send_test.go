package cmd

import (
	"io"
	"testing"
)

func TestSendRequiresFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing both flags", []string{}},
		{"missing --message", []string{"--to", "user@example.com"}},
		{"missing --to", []string{"--message", "hi"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSendCmd()
			cmd.SetArgs(tt.args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			if err := cmd.Execute(); err == nil {
				t.Error("expected error for missing required flags, got nil")
			}
		})
	}
}
