package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadPublicSSHKeysFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "builder.pub")
	input := strings.Join([]string{
		"ssh-ed25519 AAAATEST builder@example",
		"ssh-ed25519 AAAATEST builder@example",
		"not-a-key",
		"ecdsa-sha2-nistp256 BBBB second@example",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}

	keys := readPublicSSHKeysFile(path)
	if len(keys) != 2 {
		t.Fatalf("got %d keys, want 2: %v", len(keys), keys)
	}
	if keys[0] != "ssh-ed25519 AAAATEST builder@example" {
		t.Fatalf("first key = %q", keys[0])
	}
	if keys[1] != "ecdsa-sha2-nistp256 BBBB second@example" {
		t.Fatalf("second key = %q", keys[1])
	}
}

func TestReadPublicSSHKeysFileMissing(t *testing.T) {
	keys := readPublicSSHKeysFile(filepath.Join(t.TempDir(), "missing.pub"))
	if len(keys) != 0 {
		t.Fatalf("missing handoff file returned keys: %v", keys)
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
