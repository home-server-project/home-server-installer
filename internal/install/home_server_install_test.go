package install

import (
	"context"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
	"github.com/projectbluefin/knuckle/internal/runner"
)

func testHomeServerConfig(bootMiB int) *model.InstallConfig {
	return &model.InstallConfig{
		OS:                    model.OSFCOS,
		HomeServerImage:       model.HomeServerUCoreHCIImage,
		HomeServerBootSizeMiB: bootMiB,
		Channel:               "stable",
		Hostname:              "homeserver",
		Timezone:              "UTC",
		Network:               model.NetworkConfig{Mode: model.NetworkDHCP},
		Disk: model.DiskInfo{
			Path:      "/dev/disk/by-id/test-disk",
			DevPath:   "/dev/vda",
			Size:      40 * 1024 * 1024 * 1024,
			Serial:    "TESTSERIAL",
			SizeHuman: "40 GB",
		},
		Users: []model.UserConfig{{Username: "core"}},
		SSHKeys: []string{
			"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA test@key",
		},
	}
}

func TestHomeServerImageVerificationKey(t *testing.T) {
	tests := []struct {
		image   string
		wantKey string
	}{
		{model.HomeServerUCoreImage, homeServerInstallerCosignPublicKey},
		{model.HomeServerUCoreHCIImage, homeServerInstallerCosignPublicKey},
		{model.UpstreamUCoreMinimalImage, upstreamUCoreCosignPublicKey},
		{model.UpstreamUCoreImage, upstreamUCoreCosignPublicKey},
		{model.UpstreamUCoreHCIImage, upstreamUCoreCosignPublicKey},
	}
	for _, tc := range tests {
		t.Run(tc.image, func(t *testing.T) {
			got, ok := homeServerImageVerificationKey(tc.image)
			if !ok {
				t.Fatalf("expected %q to be supported", tc.image)
			}
			if got != tc.wantKey {
				t.Fatalf("verification key mismatch for %q", tc.image)
			}
		})
	}
	if _, ok := homeServerImageVerificationKey("ghcr.io/example/unknown:lts"); ok {
		t.Fatal("unknown image must not be trusted")
	}
}

func TestHomeServerInstallerAcceptsAllSupportedLTSImages(t *testing.T) {
	images := []string{
		model.HomeServerUCoreImage,
		model.HomeServerUCoreHCIImage,
		model.UpstreamUCoreMinimalImage,
		model.UpstreamUCoreImage,
		model.UpstreamUCoreHCIImage,
	}
	for _, image := range images {
		t.Run(image, func(t *testing.T) {
			spy := runner.NewSpyRunner()
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			cfg := testHomeServerConfig(model.HomeServerBootStandardMiB)
			cfg.HomeServerImage = image

			if err := NewHomeServerInstaller(spy, logger).Install(context.Background(), cfg, func(string) {}); err != nil {
				t.Fatalf("Install() error = %v", err)
			}
			if len(spy.Calls) != 1 {
				t.Fatalf("expected one direct-install invocation, got %d", len(spy.Calls))
			}
			if !strings.Contains(strings.Join(spy.Calls[0].Args, " "), image) {
				t.Fatalf("direct install args missing selected image %q", image)
			}
		})
	}
}

func TestHomeServerInstallerUsesDirectBootcScriptAndSelectedLayout(t *testing.T) {
	for _, bootMiB := range []int{model.HomeServerBootStandardMiB, model.HomeServerBootLargeMiB} {
		t.Run(strconv.Itoa(bootMiB), func(t *testing.T) {
			spy := runner.NewSpyRunner()
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			installer := NewHomeServerInstaller(spy, logger)
			cfg := testHomeServerConfig(bootMiB)

			if err := installer.Install(context.Background(), cfg, func(string) {}); err != nil {
				t.Fatalf("Install() error = %v", err)
			}
			if len(spy.Calls) != 1 {
				t.Fatalf("expected one shell invocation, got %d", len(spy.Calls))
			}
			call := spy.Calls[0]
			if call.Name != "bash" {
				t.Fatalf("expected bash runner, got %q", call.Name)
			}
			joinedArgs := strings.Join(call.Args, " ")
			if !strings.Contains(joinedArgs, cfg.Disk.DevPath) || !strings.Contains(joinedArgs, cfg.HomeServerImage) {
				t.Fatalf("direct install args missing disk/image: %s", joinedArgs)
			}
			wantBoot := strconv.Itoa(bootMiB)
			if !strings.Contains(joinedArgs, " "+wantBoot+" ") {
				t.Fatalf("args missing selected /boot size %s: %s", wantBoot, joinedArgs)
			}
			for _, required := range []string{
				"sgdisk --zap-all",
				"--typecode=3:EA00",
				"--typecode=4:8304",
				"root filesystem UUID is empty",
				"REGISTRIES_DIR=/etc/containers/registries.d",
				"00-home-server-installer.XXXXXX.yaml",
				"ghcr.io/home-server-project/home-server-ucore:",
				"ghcr.io/home-server-project/home-server-ucore-hci:",
				"ghcr.io/ublue-os/ucore-minimal:",
				"ghcr.io/ublue-os/ucore:",
				"ghcr.io/ublue-os/ucore-hci:",
				"use-sigstore-attachments: true",
				"podman pull --signature-policy",
				"rm -f -- \"$REGISTRIES_FILE\"",
				"bootc install to-filesystem",
				"--root-mount-spec",
				"--boot-mount-spec",
				"useradd --root \"$DEPLOY\"",
				"SUDOERS_FILE=\"${DEPLOY}/etc/sudoers.d/90-home-server-admin\"",
				"NOPASSWD: ALL",
				"PRESET_FILE=\"${DEPLOY}/etc/systemd/system-preset/00-home-server.preset\"",
				"disable zincati.service",
				"enable rpm-ostreed-automatic.timer",
				"systemctl --root=\"$DEPLOY\" preset rpm-ostreed-automatic.timer",
				"Home Server update preset did not survive bootc finalize",
				"PERSISTENT_VAR_ROOT=\"${TARGET_ROOT}/ostree/deploy/fedora-coreos/var\"",
				"install -d -m0755 \"$PERSISTENT_HOME_ROOT\"",
				"HOME_ROOT_CONTEXT=\"$(matchpathcon -n \"/var/home\")\"",
				"persistent /var/home has wrong SELinux context",
				"authorized_keys was not written to persistent user home",
				"bootc install finalize",
				"systemctl --root=\"$DEPLOY\" mask zincati.service",
				"zincati is not masked in finalized target",
				"rpm-ostreed-automatic.timer is not enabled in finalized target",
				"SSH-only admin sudoers file did not survive bootc finalize",
				"obsolete first-boot provisioning service remains in target",
			} {
				if !strings.Contains(call.Input, required) {
					t.Fatalf("direct install script missing %q", required)
				}
			}

			finalizePos := strings.Index(call.Input, "bootc install finalize \"$TARGET_ROOT\"")
			presetPos := strings.LastIndex(call.Input, "systemctl --root=\"$DEPLOY\" preset rpm-ostreed-automatic.timer")
			if finalizePos == -1 || presetPos == -1 || presetPos < finalizePos {
				t.Fatalf("rpm-ostree timer preset must be applied after bootc finalize")
			}

			registriesCreatePos := strings.Index(call.Input, "00-home-server-installer.XXXXXX.yaml")
			pullPos := strings.Index(call.Input, "podman pull --signature-policy")
			registriesRemovePos := strings.LastIndex(call.Input, "rm -f -- \"$REGISTRIES_FILE\"")
			if registriesCreatePos == -1 || pullPos == -1 || registriesRemovePos == -1 || !(registriesCreatePos < pullPos && pullPos < registriesRemovePos) {
				t.Fatalf("sigstore attachment config must wrap only the signed image pull")
			}

			for _, forbidden := range []string{
				"persistent target /var/home is missing",
				"home-server-provision-user.service\n[Unit]",
				"ExecStart=/etc/home-server-installer/provision-user.sh",
				"systemctl --root=\"$DEPLOY\" enable home-server-provision-user.service",
			} {
				if strings.Contains(call.Input, forbidden) {
					t.Fatalf("direct install script still contains obsolete first-boot mechanism %q", forbidden)
				}
			}
		})
	}
}

func TestHomeServerInstallerPasswordBackedAdminDoesNotGetNOPASSWD(t *testing.T) {
	spy := runner.NewSpyRunner()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := testHomeServerConfig(model.HomeServerBootStandardMiB)
	cfg.Users[0].PasswordHash = "$6$test$hash"

	if err := NewHomeServerInstaller(spy, logger).Install(context.Background(), cfg, func(string) {}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if len(spy.Calls) != 1 {
		t.Fatalf("expected one shell invocation, got %d", len(spy.Calls))
	}
	if !strings.Contains(spy.Calls[0].Input, "password-backed admin unexpectedly has passwordless sudo rule") {
		t.Fatalf("expected password-backed sudo guard in direct install script")
	}
}

func TestHomeServerInstallerDefaultsBootToOneGiB(t *testing.T) {
	spy := runner.NewSpyRunner()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := testHomeServerConfig(0)

	if err := NewHomeServerInstaller(spy, logger).Install(context.Background(), cfg, func(string) {}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if len(spy.Calls) != 1 || !strings.Contains(strings.Join(spy.Calls[0].Args, " "), " 1024 ") {
		t.Fatalf("expected default 1024 MiB layout, calls=%v", spy.Calls)
	}
}

func TestHomeServerInstallerRejectsUnknownBootSize(t *testing.T) {
	spy := runner.NewSpyRunner()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := testHomeServerConfig(1536)

	err := NewHomeServerInstaller(spy, logger).Install(context.Background(), cfg, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "unsupported Home Server /boot size") {
		t.Fatalf("expected unsupported boot size error, got %v", err)
	}
	if len(spy.Calls) != 0 {
		t.Fatalf("destructive runner must not be called for invalid layout: %v", spy.Calls)
	}
}

func TestHomeServerInstallerRejectsUnknownImage(t *testing.T) {
	spy := runner.NewSpyRunner()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := testHomeServerConfig(model.HomeServerBootStandardMiB)
	cfg.HomeServerImage = "ghcr.io/example/unknown:lts"

	err := NewHomeServerInstaller(spy, logger).Install(context.Background(), cfg, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "unsupported Home Server image") {
		t.Fatalf("expected unsupported image error, got %v", err)
	}
	if len(spy.Calls) != 0 {
		t.Fatalf("destructive runner must not be called for unsupported image: %v", spy.Calls)
	}
}

func TestFCOSInstallerRoutesHomeServerAwayFromCoreOSInstaller(t *testing.T) {
	spy := runner.NewSpyRunner()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := testHomeServerConfig(model.HomeServerBootStandardMiB)

	if err := NewFCOSInstaller(spy, logger).Install(context.Background(), cfg, func(string) {}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if len(spy.Calls) != 1 || spy.Calls[0].Name != "bash" {
		t.Fatalf("Home Server should route to direct engine, calls=%v", spy.Calls)
	}
}
