package config

import (
	"os"
	"path/filepath"
	"testing"
)

func useTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func TestLoadMissingFile(t *testing.T) {
	useTempHome(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v; want nil", err)
	}
	if cfg.ClientID != "" {
		t.Errorf("ClientID = %q; want empty", cfg.ClientID)
	}
}

func TestSaveAndLoad(t *testing.T) {
	useTempHome(t)
	want := &Config{ClientID: "test-client-id-abc"}
	if err := Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.ClientID != want.ClientID {
		t.Errorf("ClientID = %q; want %q", got.ClientID, want.ClientID)
	}
}

func TestConfigFilePermissions(t *testing.T) {
	useTempHome(t)
	if err := Save(&Config{ClientID: "x"}); err != nil {
		t.Fatal(err)
	}
	dir, _ := Dir()
	info, err := os.Stat(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("config.json permissions = %04o; want 0600", perm)
	}
}
