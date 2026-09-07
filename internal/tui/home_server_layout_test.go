package tui

import (
	"strings"
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
)

func TestHomeServerUserFormDefaultsBootLayoutToOneGiB(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepUser
	w.State.Config.OS = model.OSFCOS
	w.State.Config.HomeServerImage = model.HomeServerUCoreHCIImage
	w.State.Config.HomeServerBootSizeMiB = 0

	m := New(w)
	if m.activeForm == nil {
		t.Fatal("expected Home Server user form")
	}
	if got := w.State.Config.HomeServerBootSizeMiB; got != model.HomeServerBootStandardMiB {
		t.Fatalf("expected default %d MiB /boot, got %d", model.HomeServerBootStandardMiB, got)
	}
}

func TestHomeServerUserFormPreservesTwoGiBOverride(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepUser
	w.State.Config.OS = model.OSFCOS
	w.State.Config.HomeServerImage = model.HomeServerUCoreHCIImage
	w.State.Config.HomeServerBootSizeMiB = model.HomeServerBootLargeMiB

	_ = New(w)
	if got := w.State.Config.HomeServerBootSizeMiB; got != model.HomeServerBootLargeMiB {
		t.Fatalf("expected 2 GiB override to survive form initialization, got %d", got)
	}
}

func TestHomeServerReviewShowsChosenPartitionLayout(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepReview
	w.State.Config.OS = model.OSFCOS
	w.State.Config.HomeServerImage = model.HomeServerUCoreHCIImage
	w.State.Config.HomeServerBootSizeMiB = model.HomeServerBootLargeMiB
	w.State.Config.Disk = model.DiskInfo{DevPath: "/dev/vda", Model: "QEMU", SizeHuman: "40 GB"}
	w.State.Config.Hostname = "homeserver"
	w.State.Config.Users = []model.UserConfig{{Username: "core"}}

	m := New(w)
	summary := m.homeServerReviewSummary()
	for _, want := range []string{
		"/boot: 2 GiB XBOOTLDR (ext4)",
		"EFI: 512 MiB",
		"Root: remaining disk (XFS)",
		model.HomeServerUCoreHCIImage,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("review summary missing %q: %s", want, summary)
		}
	}
}
