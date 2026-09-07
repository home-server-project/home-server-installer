package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/projectbluefin/knuckle/internal/model"
)

func TestHomeServerInstallAndDoneNames(t *testing.T) {
	tests := []struct {
		image string
		name  string
	}{
		{model.HomeServerUCoreImage, "Home Server uCore LTS"},
		{model.HomeServerUCoreHCIImage, "Home Server uCore HCI LTS"},
		{model.UpstreamUCoreMinimalImage, "uCore Minimal LTS"},
		{model.UpstreamUCoreImage, "uCore LTS"},
		{model.UpstreamUCoreHCIImage, "uCore HCI LTS"},
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

func TestHomeServerWelcomeRestoresEachImageSelection(t *testing.T) {
	for wantCursor, opt := range homeServerImageOptions {
		w := newTestWizard()
		w.State.CurrentStep = model.StepWelcome
		w.State.Config.HomeServerImage = opt.id
		m := New(w)
		if m.cursor != wantCursor {
			t.Fatalf("image %q: expected cursor %d, got %d", opt.id, wantCursor, m.cursor)
		}
	}
}

func TestHomeServerWelcomeExposesFiveLTSChoices(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	m := New(w)

	if got := m.maxCursor(); got != 5 {
		t.Fatalf("expected five Home Server image choices, got %d", got)
	}
	out := m.viewOSPicker()
	for _, name := range []string{
		"Home Server uCore LTS",
		"Home Server uCore HCI LTS",
		"uCore Minimal LTS",
		"uCore LTS",
		"uCore HCI LTS",
	} {
		if !strings.Contains(out, name) {
			t.Fatalf("picker missing %q: %s", name, out)
		}
	}
	if !strings.Contains(out, "Switch to stable/NVIDIA later with bootc") {
		t.Fatalf("picker missing post-install rebase guidance: %s", out)
	}
}

func TestHomeServerWelcomeKeyboardSelectsUpstreamHCI(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	m := New(w)

	if m.cursor != 0 {
		t.Fatalf("expected initial Home Server cursor 0, got %d", m.cursor)
	}

	for range 4 {
		updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		m = updated.(*Model)
	}
	if m.cursor != 4 {
		t.Fatalf("expected down keys to highlight upstream HCI cursor 4, got %d", m.cursor)
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(*Model)

	if m.Wizard.State.Config.HomeServerImage != model.UpstreamUCoreHCIImage {
		t.Fatalf("expected keyboard selection to preserve upstream HCI image, got %q", m.Wizard.State.Config.HomeServerImage)
	}
	if m.Wizard.State.CurrentStep != model.StepNetwork {
		t.Fatalf("expected upstream HCI selection to advance to Network, got %v", m.Wizard.State.CurrentStep)
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
