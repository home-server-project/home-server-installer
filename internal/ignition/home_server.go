package ignition

import (
	"fmt"

	"github.com/projectbluefin/knuckle/internal/model"
)

const homeServerPublicKeyPath = "/etc/containers/home-server-project.pub"

const homeServerPublicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEWU3SeANKBm2Dql6FGZYNu2Bd7nZf
wSS/hmdv0B25JOSqi0dbyvW8XAJHJ4UOl/GeOSQM4XuDey9yI9I09r9XWw==
-----END PUBLIC KEY-----`

// GenerateHomeServerFCOSButane creates the intentionally small bootstrap
// configuration used by Home Server Installer V1. Fedora CoreOS is only an
// installation trampoline: user/network identity is provisioned, the Home
// Server signing key/policy is installed, and first boot performs a signed
// rpm-ostree rebase to uCore.
func (g *Generator) GenerateHomeServerFCOSButane(cfg *model.InstallConfig) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config cannot be nil")
	}
	if !isHomeServerImage(cfg.HomeServerImage) {
		return "", fmt.Errorf("unsupported Home Server image %q", cfg.HomeServerImage)
	}

	// Do not let generic temporary-FCOS policy leak into the final uCore
	// deployment. V1 keeps only identity/network/authentication/timezone.
	minimal := *cfg
	minimal.Sysexts = nil
	minimal.Swap = model.SwapConfig{}
	minimal.Tailscale = model.TailscaleConfig{}
	minimal.UpdateStrategy = model.UpdateStrategy{}
	minimal.NvidiaDriverVersion = ""

	b := NewBuilder(&minimal)
	if err := addSharedFragments(b, &minimal); err != nil {
		return "", err
	}

	keyFrag, err := renderTemplate("home-server-key", homeServerKeyTemplate, struct{ Key string }{Key: homeServerPublicKey})
	if err != nil {
		return "", fmt.Errorf("rendering Home Server signing key: %w", err)
	}
	b.AddStorageFile(keyFrag)

	policyFrag, err := renderTemplate("home-server-policy", homeServerPolicyTemplate, struct{ KeyPath string }{KeyPath: homeServerPublicKeyPath})
	if err != nil {
		return "", fmt.Errorf("rendering Home Server signature policy: %w", err)
	}
	b.AddStorageFile(policyFrag)

	serviceFrag, err := renderTemplate("home-server-autorebase", homeServerAutorebaseTemplate, struct{ Image string }{Image: cfg.HomeServerImage})
	if err != nil {
		return "", fmt.Errorf("rendering Home Server autorebase service: %w", err)
	}

	// Zincati belongs to the temporary FCOS deployment, not the final Home
	// Server update policy. Mask it so it cannot race the one-shot rebase.
	b.AddSystemdUnit("- name: zincati.service\n  mask: true")
	b.AddSystemdUnit(serviceFrag)

	return b.BuildFCOS(), nil
}

func isHomeServerImage(image string) bool {
	switch image {
	case model.HomeServerUCoreImage, model.HomeServerUCoreHCIImage:
		return true
	default:
		return false
	}
}

var homeServerKeyTemplate = `- path: /etc/containers/home-server-project.pub
  mode: 0644
  overwrite: true
  contents:
    inline: |
{{.Key | indentBlock 6}}`

var homeServerPolicyTemplate = `- path: /etc/containers/policy.json
  mode: 0644
  overwrite: true
  contents:
    inline: |
      {
        "default": [{"type": "insecureAcceptAnything"}],
        "transports": {
          "docker": {
            "ghcr.io/home-server-project/home-server-ucore": [{
              "type": "sigstoreSigned",
              "keyPath": "{{.KeyPath}}",
              "signedIdentity": {"type": "matchRepository"}
            }],
            "ghcr.io/home-server-project/home-server-ucore-hci": [{
              "type": "sigstoreSigned",
              "keyPath": "{{.KeyPath}}",
              "signedIdentity": {"type": "matchRepository"}
            }]
          }
        }
      }`

var homeServerAutorebaseTemplate = `- name: home-server-autorebase.service
  enabled: true
  contents: |
    [Unit]
    Description=Home Server signed uCore auto-rebase
    ConditionPathExists=!/etc/home-server-autorebase.done
    After=network-online.target
    Wants=network-online.target

    [Service]
    Type=oneshot
    StandardOutput=journal+console
    ExecStart=/usr/bin/rpm-ostree rebase --bypass-driver ostree-image-signed:docker://{{.Image}}
    ExecStart=/usr/bin/touch /etc/home-server-autorebase.done
    ExecStart=/usr/bin/systemctl disable home-server-autorebase.service
    ExecStart=/usr/bin/systemctl reboot

    [Install]
    WantedBy=multi-user.target`
