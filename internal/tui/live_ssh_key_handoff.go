package tui

import (
	"os"
	"path/filepath"
	"strings"
)

const liveInstallerBootstrapPath = "/opt/home-server-installer-bootstrap"

var liveInstallerSSHKeySources = []string{
	"/var/home/core/.ssh/authorized_keys",
	"/home/core/.ssh/authorized_keys",
	"/var/home/core/.ssh/authorized_keys.d/ignition",
	"/home/core/.ssh/authorized_keys.d/ignition",
}

// init bridges the public key embedded by build-fcos-iso.sh into Knuckle's
// existing automatic ~/.ssh/*.pub discovery. The bootstrap marker only exists
// in the live installer, so this never runs on the installed target system.
func init() {
	_ = handoffLiveInstallerSSHKey(liveInstallerBootstrapPath, liveInstallerSSHKeySources, "")
}

// handoffLiveInstallerSSHKey copies public keys from the live Fedora CoreOS
// core account into a root-owned .pub file. Knuckle runs as root in the live
// ISO and already includes ~/.ssh/*.pub files automatically, so the builder's
// --ssh-key becomes part of cfg.SSHKeys without a second paste step.
//
// Only public keys are copied. The destination is live-installer state and is
// never written directly to the installed system; the Home Server installer
// later provisions the selected user's authorized_keys through its normal
// persistence path.
func handoffLiveInstallerSSHKey(bootstrapPath string, sourcePaths []string, homeOverride string) error {
	if _, err := os.Stat(bootstrapPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	home := homeOverride
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return err
		}
	}

	for _, sourcePath := range sourcePaths {
		data, err := os.ReadFile(sourcePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		keys := publicSSHKeyLines(string(data))
		if len(keys) == 0 {
			continue
		}

		sshDir := filepath.Join(home, ".ssh")
		if err := os.MkdirAll(sshDir, 0o700); err != nil {
			return err
		}
		if err := os.Chmod(sshDir, 0o700); err != nil {
			return err
		}

		destination := filepath.Join(sshDir, "home-server-installer.pub")
		contents := strings.Join(keys, "\n") + "\n"
		if err := os.WriteFile(destination, []byte(contents), 0o600); err != nil {
			return err
		}
		return os.Chmod(destination, 0o600)
	}

	return nil
}

func publicSSHKeyLines(contents string) []string {
	seen := map[string]struct{}{}
	var keys []string
	for _, line := range strings.Split(contents, "\n") {
		key := strings.TrimSpace(line)
		if !(strings.HasPrefix(key, "ssh-") || strings.HasPrefix(key, "ecdsa-") || strings.HasPrefix(key, "sk-")) {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return keys
}
