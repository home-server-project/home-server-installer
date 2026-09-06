package tui

import (
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
)

func TestHomeServerReviewNoReturnsToInitializedUserForm(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepReview
	w.State.Config.OS = model.OSFCOS
	w.State.Config.Channel = "stable"
	w.State.Config.HomeServerImage = model.HomeServerUCoreHCIImage
	w.State.Confirmed = false

	m := New(w)
	cmd := m.onFormComplete()

	if m.Wizard.State.CurrentStep != model.StepUser {
		t.Fatalf("expected Home Server review back to land on User, got %v", m.Wizard.State.CurrentStep)
	}
	if m.activeForm == nil {
		t.Fatal("expected User form to be restored after review back")
	}
	if cmd == nil {
		t.Fatal("expected restored User form to return its Init command")
	}
	if m.Wizard.State.Config.HomeServerImage != model.HomeServerUCoreHCIImage {
		t.Fatalf("expected HCI target to survive review back, got %q", m.Wizard.State.Config.HomeServerImage)
	}
}
