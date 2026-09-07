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
		"sk-ssh-ed25519@openssh.com CCCC three@example",
		"not-a-key",
		"",
	}, "\n")

	keys := publicSSHKeyLines(input)
	if len(keys) != 3 {
		t.Fatalf("got %d keys, want 3: %v", len(keys), keys)
	}
	if keys[0] != "ssh-ed25519 AAAA one@example" {
		t.Fatalf("first key = %q", keys[0])
	}
	if keys[1] != "ecdsa-sha2-nistp256 BBBB two@example" {
		t.Fatalf("second key = %q", keys[1])
	}
	if keys[2] != "sk-ssh-ed25519@openssh.com CCCC three@example" {
		t.Fatalf("third key = %q", keys[2])
	}
}

func TestHomeServerKeysSummaryFor(t *testing.T) {
	tests := []struct {
		name        string
		builderKeys []string
		localKeys   []string
		want        string
	}{
		{
			name:        "builder and local",
			builderKeys: []string{"ssh-ed25519 AAAA builder"},
			localKeys:   []string{"ssh-ed25519 BBBB local1", "ssh-ed25519 CCCC local2"},
			want:        "Builder SSH key detected; 2 additional local key(s) will also be included automatically",
		},
		{
			name:        "builder only",
			builderKeys: []string{"ssh-ed25519 AAAA builder"},
			want:        "Builder SSH key detected — it will be installed automatically",
		},
		{
			name:      "local only",
			localKeys: []string{"ssh-ed25519 BBBB local"},
			want:      "1 local key(s) from ~/.ssh/ will be included automatically",
		},
		{
			name: "none",
			want: "No automatic SSH keys detected; add a GitHub username or paste a public key if desired",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := homeServerKeysSummaryFor(tc.builderKeys, tc.localKeys)
			if !strings.Contains(got, tc.want) {
				t.Fatalf("summary = %q, want substring %q", got, tc.want)
			}
		})
	}
}
