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

func TestHomeServerNvidiaImagesUseUpstreamVerificationKey(t *testing.T) {
	images := []string{
		model.UpstreamUCoreMinimalNvidiaImage,
		model.UpstreamUCoreNvidiaImage,
		model.UpstreamUCoreHCINvidiaImage,
		model.UpstreamUCoreMinimalNvidiaLTSImage,
		model.UpstreamUCoreNvidiaLTSImage,
		model.UpstreamUCoreHCINvidiaLTSImage,
	}

	for _, image := range images {
		t.Run(image, func(t *testing.T) {
			key, ok := homeServerImageVerificationKey(image)
			if !ok {
				t.Fatalf("expected %q to be supported", image)
			}
			if key != upstreamUCoreCosignPublicKey {
				t.Fatalf("expected upstream uCore verification key for %q", image)
			}
		})
	}
}

func TestHomeServerInstallerAcceptsNvidiaLTSImages(t *testing.T) {
	images := []string{
		model.UpstreamUCoreMinimalNvidiaImage,
		model.UpstreamUCoreNvidiaImage,
		model.UpstreamUCoreHCINvidiaImage,
		model.UpstreamUCoreMinimalNvidiaLTSImage,
		model.UpstreamUCoreNvidiaLTSImage,
		model.UpstreamUCoreHCINvidiaLTSImage,
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
