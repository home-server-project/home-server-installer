package tui

import (
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
)

func TestHomeServerWelcomeOutOfRangeCursorFallsBackToUCore(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	m := New(w)
	m.cursor = 99

	_, _ = m.handleEnter()

	if m.Wizard.State.Config.HomeServerImage != model.HomeServerUCoreImage {
		t.Fatalf("expected out-of-range cursor to fall back to standard uCore, got %q", m.Wizard.State.Config.HomeServerImage)
	}
	if m.Wizard.State.CurrentStep != model.StepNetwork {
		t.Fatalf("expected StepNetwork, got %v", m.Wizard.State.CurrentStep)
	}
}

func TestPreservedGenericWelcomeCanStillSelectFCOSStreamAndIgnition(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	w.State.Config.OS = model.OSFCOS
	w.State.Config.Channel = "stable"
	w.State.Config.IgnitionURL = "https://example.com/config.ign"

	m := New(w)
	// The Home Server profile hides this generic Knuckle view, but we keep the
	// implementation intact for upstream sync and future image choices.
	m.osSubView = false
	m.cursor = 1 // testing

	_, _ = m.handleEnter()

	if m.Wizard.State.Config.Channel != "testing" {
		t.Fatalf("expected preserved FCOS stream selection to set testing, got %q", m.Wizard.State.Config.Channel)
	}
	if m.Wizard.State.CurrentStep != model.StepStorage {
		t.Fatalf("expected external Ignition path to skip to Storage, got %v", m.Wizard.State.CurrentStep)
	}
}

func TestHomeServerReviewFormCoversBothTargets(t *testing.T) {
	for _, image := range []string{
		model.HomeServerUCoreImage,
		model.HomeServerUCoreHCIImage,
	} {
		t.Run(image, func(t *testing.T) {
			w := newTestWizard()
			w.State.CurrentStep = model.StepReview
			w.State.Config.OS = model.OSFCOS
			w.State.Config.Channel = "stable"
			w.State.Config.HomeServerImage = image
			w.State.Config.Disk = model.DiskInfo{DevPath: "/dev/vda", Model: "QEMU", SizeHuman: "40G"}
			m := New(w)

			if m.activeForm == nil {
				t.Fatal("expected review confirmation form")
			}
			if summary := m.reviewSummary(); summary == "" {
				t.Fatal("expected non-empty review summary")
			}
		})
	}
}
