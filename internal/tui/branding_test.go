package tui

import (
	"strings"
	"testing"
)

func TestRenderZenChromeHomeServerBranding(t *testing.T) {
	w := newTestWizard()
	m := New(w)
	out := m.renderZenChrome()

	for _, want := range []string{
		"Home Server Project · based on Project Bluefin Knuckle",
		"H O M E   S E R V E R   I N S T A L L E R",
		"Friendly installer for Home Server Gina and Universal Blue uCore LTS",
		"Cloud-native technology, brought home.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("branding output should contain %q: %q", want, out)
		}
	}

	for _, old := range []string{
		"K N U C K L E",
		"Home Server Gina installer",
		"The real thing, right from the CNCF. Legends will rise.",
		"powered by Project Bluefin Knuckle",
	} {
		if strings.Contains(out, old) {
			t.Errorf("branding output should not contain old text %q: %q", old, out)
		}
	}
}

func TestViewOSPickerUsesNeutralInstallationWording(t *testing.T) {
	w := newTestWizard()
	m := New(w)
	out := m.viewOSPicker()

	for _, want := range []string{
		"Select installation target:",
		"Home Server Gina LTS",
		"Home Server Gina HCI LTS",
		"uCore Minimal LTS",
		"uCore LTS",
		"uCore HCI LTS",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("target picker should contain %q: %q", want, out)
		}
	}

	for _, old := range []string{
		"Select Home Server image:",
		"Recommended Home Server Project image",
	} {
		if strings.Contains(out, old) {
			t.Errorf("target picker should not contain old text %q: %q", old, out)
		}
	}
}
