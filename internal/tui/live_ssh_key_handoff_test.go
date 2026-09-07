package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandoffLiveInstallerSSHKey(t *testing.T) {
	tmp := t.TempDir()
	bootstrap := filepath.Join(tmp, "home-server-installer-bootstrap")
	if err := os.WriteFile(bootstrap, []byte("marker\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	source := filepath.Join(tmp, "authorized_keys")
	key := "ssh-ed25519 AAAATEST builder@example"
	if err := os.WriteFile(source, []byte(key+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	home := filepath.Join(tmp, "root")
	if err := handoffLiveInstallerSSHKey(bootstrap, []string{source}, home); err != nil {
		t.Fatalf("handoffLiveInstallerSSHKey: %v", err)
	}

	destination := filepath.Join(home, ".ssh", "home-server-installer.pub")
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("reading destination: %v", err)
	}
	if string(data) != key+"\n" {
		t.Fatalf("destination = %q, want %q", string(data), key+"\n")
	}

	info, err := os.Stat(destination)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("destination mode = %o, want 600", info.Mode().Perm())
	}
}

func TestHandoffLiveInstallerSSHKeyRequiresLiveBootstrap(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "authorized_keys")
	if err := os.WriteFile(source, []byte("ssh-ed25519 AAAATEST builder@example\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	home := filepath.Join(tmp, "root")
	missingBootstrap := filepath.Join(tmp, "missing-bootstrap")
	if err := handoffLiveInstallerSSHKey(missingBootstrap, []string{source}, home); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(home, ".ssh", "home-server-installer.pub")); !os.IsNotExist(err) {
		t.Fatalf("builder key should not be created without live bootstrap marker, err=%v", err)
	}
}

func TestPublicSSHKeyLinesFiltersAndDeduplicates(t *testing.T) {
	input := strings.Join([]string{
		"# comment",
		"ssh-ed25519 AAAA one@example",
		"ssh-ed25519 AAAA one@example",
		"ecdsa-sha2-nistp256 BBBB two@example",
		"not-a-key",
		"",
	}, "\n")

	keys := publicSSHKeyLines(input)
	if len(keys) != 2 {
		t.Fatalf("got %d keys, want 2: %v", len(keys), keys)
	}
	if keys[0] != "ssh-ed25519 AAAA one@example" {
		t.Fatalf("first key = %q", keys[0])
	}
	if keys[1] != "ecdsa-sha2-nistp256 BBBB two@example" {
		t.Fatalf("second key = %q", keys[1])
	}
}
