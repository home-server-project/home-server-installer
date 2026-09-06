package tui

import (
	"strings"
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
)

func TestHomeServerInstallAndDoneNames(t *testing.T) {
	tests := []struct {
		image string
		name  string
	}{
		{model.HomeServerUCoreImage, "Home Server uCore"},
		{model.HomeServerUCoreHCIImage, "Home Server uCore HCI"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newTestWizard()
			w.State.Config.OS = model.OSFCOS
			w.State.Config.Channel = "stable"
			w.State.Config.HomeServerImage = tc.image
			m := New(w)

			if out := m.viewInstall(); !strings.Contains(out, "Installing "+tc.name) {
				t.Fatalf("install view missing %q: %s", tc.name, out)
			}
			if out := m.viewDone(); !strings.Contains(out, tc.name+" has been installed") {
				t.Fatalf("done view missing %q: %s", tc.name, out)
			}
		})
	}
}

func TestHomeServerWelcomeRestoresHCISelection(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	w.State.Config.HomeServerImage = model.HomeServerUCoreHCIImage
	m := New(w)
	if m.cursor != 1 {
		t.Fatalf("expected HCI cursor 1, got %d", m.cursor)
	}
}

func TestHomeServerUserFormDefaultsHostname(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepUser
	w.State.Config.HomeServerImage = model.HomeServerUCoreImage
	w.State.Config.Hostname = ""
	_ = New(w)
	if w.State.Config.Hostname != "homeserver" {
		t.Fatalf("expected homeserver default hostname, got %q", w.State.Config.Hostname)
	}
}
