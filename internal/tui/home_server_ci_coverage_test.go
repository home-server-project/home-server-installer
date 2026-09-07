package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
)

const homeServerCoveragePublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA coverage@test"

func TestHomeServerUserCompletionIncludesAutomaticLocalKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_ed25519.pub"), []byte(homeServerCoveragePublicKey+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	w := newTestWizard()
	w.State.CurrentStep = model.StepUser
	w.State.Config.OS = model.OSFCOS
	w.State.Config.Channel = "stable"
	w.State.Config.HomeServerImage = model.HomeServerUCoreImage
	w.State.Config.Hostname = "homeserver"
	w.State.Config.Timezone = "UTC"

	m := New(w)
	m.usernameInput = "core"
	m.passwordInput = ""
	m.githubUserInput = ""
	m.sshKeyInput = ""

	_ = m.onFormComplete()

	if m.err != nil {
		t.Fatalf("unexpected Home Server user-step error: %v", m.err)
	}
	if !containsString(m.Wizard.State.Config.SSHKeys, homeServerCoveragePublicKey) {
		t.Fatalf("automatic local key was not included: %v", m.Wizard.State.Config.SSHKeys)
	}
}

func TestHomeServerAutomaticSSHKeysIncludesLocalDiscovery(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "workstation.pub"), []byte(homeServerCoveragePublicKey+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	keys := homeServerAutomaticSSHKeys()
	if !containsString(keys, homeServerCoveragePublicKey) {
		t.Fatalf("homeServerAutomaticSSHKeys() = %v, want local key", keys)
	}
}

func TestHomeServerReviewSummaryDefaultsBootToOneGiB(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepReview
	w.State.Config.OS = model.OSFCOS
	w.State.Config.Channel = "stable"
	w.State.Config.HomeServerImage = model.HomeServerUCoreImage
	w.State.Config.HomeServerBootSizeMiB = 0
	w.State.Config.Disk = model.DiskInfo{DevPath: "/dev/vda", Model: "QEMU", SizeHuman: "40 GB"}
	w.State.Config.Users = []model.UserConfig{{Username: "core"}}

	m := New(w)
	summary := m.homeServerReviewSummary()
	if !strings.Contains(summary, "/boot: 1 GiB XBOOTLDR (ext4)") {
		t.Fatalf("default Home Server review summary missing 1 GiB layout: %s", summary)
	}
}

func TestHomeServerReviewSummaryFallbackWithoutDestination(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepReview
	w.State.Config.OS = model.OSFCOS
	w.State.Config.Channel = "stable"
	w.State.Config.HomeServerImage = ""
	w.State.Config.Disk = model.DiskInfo{DevPath: "/dev/vda"}

	m := New(w)
	summary := m.homeServerReviewSummary()
	if !strings.Contains(summary, "OS: fcos") {
		t.Fatalf("fallback review summary lost generic OS line: %s", summary)
	}
	if strings.Contains(summary, "XBOOTLDR") {
		t.Fatalf("generic fallback unexpectedly injected Home Server layout: %s", summary)
	}
}

func TestInstallTargetDisplayNameGenericFallbacks(t *testing.T) {
	tests := []struct {
		name string
		cfg  model.InstallConfig
		want string
	}{
		{name: "fcos", cfg: model.InstallConfig{OS: model.OSFCOS}, want: "Fedora CoreOS"},
		{name: "bluefin", cfg: model.InstallConfig{OS: model.OSBluefinDDI}, want: "Bluefin Server"},
		{name: "flatcar", cfg: model.InstallConfig{}, want: "Flatcar Container Linux"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := installTargetDisplayName(&tc.cfg); got != tc.want {
				t.Fatalf("installTargetDisplayName() = %q, want %q", got, tc.want)
			}
		})
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
