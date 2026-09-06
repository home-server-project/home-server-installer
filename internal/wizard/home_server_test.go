package wizard

import (
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
)

func TestHomeServerProfileSkipsGenericBootstrapSteps(t *testing.T) {
	w := New(nil, nil, nil)
	w.State.Config.OS = model.OSFCOS
	w.State.Config.HomeServerImage = model.HomeServerUCoreImage
	w.State.Config.Hostname = "homeserver"
	w.State.Config.Users = []model.UserConfig{{Username: "core", SSHKeys: []string{"ssh-ed25519 AAAATEST test"}}}
	w.State.CurrentStep = model.StepUser

	if err := w.Next(); err != nil {
		t.Fatalf("Next: %v", err)
	}
	if w.State.CurrentStep != model.StepReview {
		t.Fatalf("Home Server should skip generic FCOS steps and land on Review, got %v", w.State.CurrentStep)
	}

	w.Previous()
	if w.State.CurrentStep != model.StepUser {
		t.Fatalf("Previous should skip generic FCOS steps and return to User, got %v", w.State.CurrentStep)
	}
}
