package tui

import (
	"fmt"
	"os"
	"strings"
)

const liveInstallerSSHKeyPath = "/opt/home-server-installer-ssh.pub"

// detectLiveInstallerBuilderSSHKeys reads the dedicated public-key handoff
// created by build-fcos-iso.sh when --ssh-key is supplied. The file exists only
// in the live Fedora CoreOS installer environment; it is not part of the bootc
// target filesystem.
func detectLiveInstallerBuilderSSHKeys() []string {
	return readPublicSSHKeysFile(liveInstallerSSHKeyPath)
}

func readPublicSSHKeysFile(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return publicSSHKeyLines(string(data))
}

// homeServerAutomaticSSHKeys combines normal local *.pub discovery with the
// explicit builder handoff. The builder path does not depend on HOME, which is
// intentionally not set for the root-owned installer systemd service.
func homeServerAutomaticSSHKeys() []string {
	return mergeKeys(detectLocalSSHKeys(), detectLiveInstallerBuilderSSHKeys())
}

func (m *Model) homeServerKeysSummary() string {
	builderKeys := detectLiveInstallerBuilderSSHKeys()
	localKeys := detectLocalSSHKeys()

	switch {
	case len(builderKeys) > 0 && len(localKeys) > 0:
		return fmt.Sprintf("🔑 Builder SSH key detected; %d additional local key(s) will also be included automatically", len(localKeys))
	case len(builderKeys) > 0:
		return "🔑 Builder SSH key detected — it will be installed automatically"
	case len(localKeys) > 0:
		return fmt.Sprintf("🔑 %d local key(s) from ~/.ssh/ will be included automatically", len(localKeys))
	default:
		return "⚠ No automatic SSH keys detected; add a GitHub username or paste a public key if desired"
	}
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
