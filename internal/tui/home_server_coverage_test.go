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
		{model.HomeServerUCoreImage, "Home Server Gina LTS"},
		{model.HomeServerUCoreHCIImage, "Home Server Gina HCI LTS"},
		{model.UpstreamUCoreMinimalImage, "uCore Minimal LTS"},
		{model.UpstreamUCoreImage, "uCore LTS"},
		{model.UpstreamUCoreHCIImage, "uCore HCI LTS"},
		{model.UpstreamUCoreMinimalNvidiaImage, "uCore Minimal LTS NVIDIA Open"},
		{model.UpstreamUCoreNvidiaImage, "uCore LTS NVIDIA Open"},
		{model.UpstreamUCoreHCINvidiaImage, "uCore HCI LTS NVIDIA Open"},
		{model.UpstreamUCoreMinimalNvidiaLTSImage, "uCore Minimal LTS NVIDIA LTS"},
		{model.UpstreamUCoreNvidiaLTSImage, "uCore LTS NVIDIA LTS"},
		{model.UpstreamUCoreHCINvidiaLTSImage, "uCore HCI LTS NVIDIA LTS"},
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

func TestHomeServerWelcomeRestoresImageSelection(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	w.State.Config.HomeServerImage = model.UpstreamUCoreHCIImage

	m := New(w)
	if m.homeServerFamily != homeServerFamilyUCore {
		t.Fatalf("expected uCore family restored, got %q", m.homeServerFamily)
	}
	if m.cursor != 2 {
		t.Fatalf("expected uCore HCI edition cursor 2, got %d", m.cursor)
	}
	if out := m.viewOSPicker(); !strings.Contains(out, "Select Universal Blue uCore LTS edition:") {
		t.Fatalf("restored picker should show uCore editions: %s", out)
	}
}

func TestHomeServerWelcomeExposesFourFamilies(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	m := New(w)

	if got := m.maxCursor(); got != 4 {
		t.Fatalf("expected four Home Server families, got %d", got)
	}
	out := m.viewOSPicker()
	for _, name := range []string{
		"Home Server Gina LTS",
		"Universal Blue uCore LTS",
		"NVIDIA Open",
		"NVIDIA LTS",
	} {
		if !strings.Contains(out, name) {
			t.Fatalf("family picker missing %q: %s", name, out)
		}
	}
	if strings.Contains(out, "uCore Minimal LTS") || strings.Contains(out, "Home Server Gina HCI LTS") {
		t.Fatalf("family picker should not expose editions yet: %s", out)
	}
	if !strings.Contains(out, "Switch to stable/testing later with bootc") {
		t.Fatalf("picker missing post-install rebase guidance: %s", out)
	}
}

func TestHomeServerWelcomeKeyboardSelectsUpstreamHCI(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	m := New(w)

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = updated.(*Model)
	if m.cursor != 1 {
		t.Fatalf("expected uCore family cursor 1, got %d", m.cursor)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(*Model)
	if m.homeServerFamily != homeServerFamilyUCore {
		t.Fatalf("expected uCore family selection, got %q", m.homeServerFamily)
	}
	if m.Wizard.State.Config.HomeServerImage != "" {
		t.Fatalf("family selection must not set an install image, got %q", m.Wizard.State.Config.HomeServerImage)
	}

	for range 2 {
		updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		m = updated.(*Model)
	}
	if m.cursor != 2 {
		t.Fatalf("expected uCore HCI edition cursor 2, got %d", m.cursor)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(*Model)
	if m.Wizard.State.Config.HomeServerImage != model.UpstreamUCoreHCIImage {
		t.Fatalf("expected upstream HCI image, got %q", m.Wizard.State.Config.HomeServerImage)
	}
	if m.Wizard.State.CurrentStep != model.StepNetwork {
		t.Fatalf("expected upstream HCI selection to advance to Network, got %v", m.Wizard.State.CurrentStep)
	}
}

func TestHomeServerWelcomeEscReturnsEditionToFamily(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	m := New(w)
	m.cursor = 2

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(*Model)
	if m.homeServerFamily != homeServerFamilyNvidia {
		t.Fatalf("expected NVIDIA Open family, got %q", m.homeServerFamily)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = updated.(*Model)
	if m.homeServerFamily != "" {
		t.Fatalf("Esc should return to family picker, got %q", m.homeServerFamily)
	}
	if m.cursor != 2 {
		t.Fatalf("Esc should restore selected family cursor 2, got %d", m.cursor)
	}
	if m.Wizard.State.Config.HomeServerImage != "" {
		t.Fatalf("Esc from edition picker should leave image empty, got %q", m.Wizard.State.Config.HomeServerImage)
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
