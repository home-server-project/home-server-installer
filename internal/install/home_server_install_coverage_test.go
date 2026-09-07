package install

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
	"github.com/projectbluefin/knuckle/internal/runner"
)

func testHomeServerLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHomeServerInstallerRejectsNilConfig(t *testing.T) {
	err := NewHomeServerInstaller(runner.NewSpyRunner(), testHomeServerLogger()).Install(context.Background(), nil, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "install config cannot be nil") {
		t.Fatalf("expected nil config error, got %v", err)
	}
}

func TestHomeServerInstallerRequiresPrimaryUser(t *testing.T) {
	tests := []struct {
		name  string
		users []model.UserConfig
	}{
		{name: "missing", users: nil},
		{name: "blank", users: []model.UserConfig{{Username: "   "}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testHomeServerConfig(model.HomeServerBootStandardMiB)
			cfg.Users = tc.users
			err := NewHomeServerInstaller(runner.NewSpyRunner(), testHomeServerLogger()).Install(context.Background(), cfg, func(string) {})
			if err == nil || !strings.Contains(err.Error(), "requires a primary user") {
				t.Fatalf("expected primary user error, got %v", err)
			}
		})
	}
}

func TestHomeServerInstallerNoSSHKeysAndDirectDiskPath(t *testing.T) {
	spy := runner.NewSpyRunner()
	cfg := testHomeServerConfig(model.HomeServerBootStandardMiB)
	cfg.SSHKeys = nil
	cfg.Disk.Path = cfg.Disk.DevPath

	var progress []string
	if err := NewHomeServerInstaller(spy, testHomeServerLogger()).Install(context.Background(), cfg, func(step string) {
		progress = append(progress, step)
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if len(spy.Calls) != 1 {
		t.Fatalf("expected one runner call, got %d", len(spy.Calls))
	}
	if got := spy.Calls[0].Args[3]; got != "" {
		t.Fatalf("stable path argument = %q, want empty", got)
	}
	if len(progress) != 2 || progress[0] != "Preparing Home Server disk layout..." || progress[1] != "Installation complete!" {
		t.Fatalf("unexpected progress callbacks: %v", progress)
	}
}

func TestHomeServerInstallerPropagatesRunnerError(t *testing.T) {
	spy := runner.NewSpyRunner()
	spy.AllError = errors.New("runner failed")
	cfg := testHomeServerConfig(model.HomeServerBootStandardMiB)

	err := NewHomeServerInstaller(spy, testHomeServerLogger()).Install(context.Background(), cfg, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "Home Server direct installation failed") {
		t.Fatalf("expected runner failure, got %v", err)
	}
}

func TestWritePrivateTempSuccessAndCreateFailure(t *testing.T) {
	path, err := writePrivateTemp("home-server-test-*", "secret\n")
	if err != nil {
		t.Fatalf("writePrivateTemp() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat temp file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("temp mode = %o, want 600", info.Mode().Perm())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read temp file: %v", err)
	}
	if string(data) != "secret\n" {
		t.Fatalf("temp content = %q", data)
	}

	missing := filepath.Join(t.TempDir(), "missing", "tmp")
	t.Setenv("TMPDIR", missing)
	if _, err := writePrivateTemp("home-server-test-*", "x"); err == nil {
		t.Fatal("expected CreateTemp failure with missing TMPDIR")
	}
}
