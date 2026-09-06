package ignition

import (
	"strings"
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
)

func TestGenerateHomeServerFCOSButane_SignedMinimalBootstrap(t *testing.T) {
	cfg := &model.InstallConfig{
		OS:              model.OSFCOS,
		HomeServerImage: model.HomeServerUCoreImage,
		Channel:         "stable",
		Hostname:        "homeserver",
		Timezone:        "UTC",
		SSHKeys:         []string{"ssh-ed25519 AAAATEST home@test"},
		Swap:            model.SwapConfig{Enabled: true, SizeMB: 4096},
		Sysexts: []model.SysextEntry{{
			Name:     "should-not-survive",
			URL:      "https://example.invalid/extension.raw",
			Selected: true,
		}},
		Tailscale: model.TailscaleConfig{AuthKey: "tskey-auth-example"},
	}

	out, err := NewGenerator().GenerateFCOSButane(cfg)
	if err != nil {
		t.Fatalf("GenerateFCOSButane: %v", err)
	}

	for _, want := range []string{
		"variant: fcos",
		model.HomeServerUCoreImage,
		"ostree-image-signed:docker://",
		"sigstoreSigned",
		"home-server-project.pub",
		"home-server-autorebase.service",
		"zincati.service",
		"mask: true",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"/etc/zincati/config.d/55-updates.toml",
		"/var/swapfile",
		"/etc/extensions/should-not-survive.raw",
		"/etc/tailscale/tailscale.env",
	} {
		if strings.Contains(out, unwanted) {
			t.Errorf("Home Server bootstrap unexpectedly contains %q\n%s", unwanted, out)
		}
	}

	if _, err := CompileToIgnition(out); err != nil {
		t.Fatalf("Home Server Butane did not compile: %v\n%s", err, out)
	}
}

func TestGenerateHomeServerFCOSButane_RejectsUnknownImage(t *testing.T) {
	cfg := &model.InstallConfig{OS: model.OSFCOS, HomeServerImage: "ghcr.io/example/not-home-server:latest"}
	if _, err := NewGenerator().GenerateFCOSButane(cfg); err == nil {
		t.Fatal("expected unsupported Home Server image to fail")
	}
}
