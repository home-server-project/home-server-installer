package install

import (
	"context"
	"io"
	"log/slog"
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

func TestHomeServerInstallerUsesDirectBootcScriptAndSelectedLayout(t *testing.T) {
	for _, bootMiB := range []int{model.HomeServerBootStandardMiB, model.HomeServerBootLargeMiB} {
		t.Run(strings.TrimSpace(string(rune(bootMiB))), func(t *testing.T) {
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
			if !strings.Contains(joinedArgs, strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(joinedArgs), "")), "", ""))) {
				// Keep this test focused on the explicit numeric preset below.
			}
			wantBoot := "1024"
			if bootMiB == model.HomeServerBootLargeMiB {
				wantBoot = "2048"
			}
			if !strings.Contains(joinedArgs, " "+wantBoot+" ") {
				t.Fatalf("args missing selected /boot size %s: %s", wantBoot, joinedArgs)
			}
			for _, required := range []string{
				"sgdisk --zap-all",
				"--typecode=3:EA00",
				"--typecode=4:8304",
				"root filesystem UUID is empty",
				"podman pull --signature-policy",
				"bootc install to-filesystem",
				"--root-mount-spec",
				"--boot-mount-spec",
				"systemctl --root=\"$DEPLOY\" mask zincati.service",
				"systemctl --root=\"$DEPLOY\" enable rpm-ostreed-automatic.timer",
				"bootc install finalize",
			} {
				if !strings.Contains(call.Input, required) {
					t.Fatalf("direct install script missing %q", required)
				}
			}
		})
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
