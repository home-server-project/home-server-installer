// Modified by Home Server Project from Project Bluefin Knuckle.
package tui

import (
	"strings"
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
	"github.com/projectbluefin/knuckle/internal/wizard"
)

// ── Home Server target picker ────────────────────────────────────────────────

func TestOSPicker_ShowsHomeServerFamilies(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	m := New(w)
	if !m.osSubView {
		t.Fatal("osSubView should be true after New() at StepWelcome")
	}
	out := m.viewChannelCards()
	for _, want := range []string{
		"Home Server Gina LTS",
		"Universal Blue uCore LTS",
		"NVIDIA Open",
		"NVIDIA LTS",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("family picker should show %q: %q", want, out)
		}
	}
	if strings.Contains(out, "Gina HCI LTS") {
		t.Errorf("family picker should not show Gina editions before family selection: %q", out)
	}
	if !strings.Contains(out, "All installer choices use LTS. Switch to stable/testing later with bootc.") {
		t.Errorf("family picker should show LTS guidance: %q", out)
	}
}

func TestOSPicker_SelectGinaThenEdition_AdvancesToNetwork(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	m := New(w)

	// First Enter selects the Gina family and stays on Welcome.
	m.cursor = 0
	_, _ = m.handleEnter()
	if m.Wizard.State.CurrentStep != model.StepWelcome {
		t.Fatalf("family selection should stay on StepWelcome, got %v", m.Wizard.State.CurrentStep)
	}
	out := m.viewChannelCards()
	if !strings.Contains(out, "Gina LTS") || !strings.Contains(out, "Gina HCI LTS") {
		t.Fatalf("Gina edition picker missing expected editions: %q", out)
	}

	// Second Enter selects Gina LTS and continues to Network.
	m.cursor = 0
	_, _ = m.handleEnter()
	if m.Wizard.State.Config.OS != model.OSFCOS {
		t.Errorf("expected bootstrap OS=fcos, got %q", m.Wizard.State.Config.OS)
	}
	if m.Wizard.State.Config.HomeServerImage != model.HomeServerUCoreImage {
		t.Errorf("unexpected Home Server image %q", m.Wizard.State.Config.HomeServerImage)
	}
	if m.Wizard.State.CurrentStep != model.StepNetwork {
		t.Errorf("expected StepNetwork, got %v", m.Wizard.State.CurrentStep)
	}
	if m.Wizard.State.Config.Swap.Enabled {
		t.Error("Home Server bootstrap must not provision generic FCOS swap")
	}
}

func TestOSPicker_SelectGinaHCI_AdvancesToNetwork(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	m := New(w)

	m.cursor = 0
	_, _ = m.handleEnter()
	_ = m.viewChannelCards() // synchronize Gina edition choices
	m.cursor = 1
	_, _ = m.handleEnter()

	if m.Wizard.State.Config.HomeServerImage != model.HomeServerUCoreHCIImage {
		t.Errorf("unexpected Home Server image %q", m.Wizard.State.Config.HomeServerImage)
	}
	if m.Wizard.State.CurrentStep != model.StepNetwork {
		t.Errorf("expected StepNetwork, got %v", m.Wizard.State.CurrentStep)
	}
}

func TestOSPicker_NvidiaFamiliesUseExpectedTags(t *testing.T) {
	tests := []struct {
		familyCursor int
		wantImage    string
	}{
		{2, model.UpstreamUCoreMinimalNvidiaImage},
		{3, model.UpstreamUCoreMinimalNvidiaLTSImage},
	}

	for _, tc := range tests {
		w := newTestWizard()
		w.State.CurrentStep = model.StepWelcome
		m := New(w)
		_ = m.viewChannelCards() // synchronize family choices
		m.cursor = tc.familyCursor
		_, _ = m.handleEnter()
		_ = m.viewChannelCards() // synchronize selected family's editions
		m.cursor = 0
		_, _ = m.handleEnter()
		if m.Wizard.State.Config.HomeServerImage != tc.wantImage {
			t.Errorf("family cursor %d selected %q, want %q", tc.familyCursor, m.Wizard.State.Config.HomeServerImage, tc.wantImage)
		}
	}
}

// ── FCOS stream cards ─────────────────────────────────────────────────────────

func TestFCOSStreamCards_ListsThreeStreams(t *testing.T) {
	w := newTestWizard()
	w.State.Config.OS = model.OSFCOS
	m := New(w)
	m.osSubView = false

	out := m.viewChannelCards()
	for _, stream := range []string{"Stable", "Testing", "Next"} {
		if !strings.Contains(out, stream) {
			t.Errorf("FCOS stream cards should contain %q: %q", stream, out[:min(300, len(out))])
		}
	}
}

func TestFCOSStreamCards_ChannelList(t *testing.T) {
	w := newTestWizard()
	w.State.Config.OS = model.OSFCOS
	m := New(w)

	got := m.channelList()
	want := []string{"stable", "testing", "next"}
	if len(got) != len(want) {
		t.Fatalf("FCOS channelList length = %d, want %d", len(got), len(want))
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("channelList[%d] = %q, want %q", i, got[i], v)
		}
	}
}

func TestFCOSStreamCards_WithVersionInfo(t *testing.T) {
	w := newTestWizard()
	w.State.Config.OS = model.OSFCOS
	w.State.FCOSStreams = []wizard.FCOSStreamInfo{
		{Stream: "stable", Version: "42.20250101.3.0"},
	}
	m := New(w)
	m.osSubView = false
	m.cursor = 0

	out := m.viewChannelCards()
	if !strings.Contains(out, "42.20250101.3.0") {
		t.Errorf("FCOS stream cards should show version: %q", out[:min(300, len(out))])
	}
}

// ── FCOS maxCursor ────────────────────────────────────────────────────────────

func TestMaxCursor_FCOSUpdate(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepUpdate
	w.State.Config.OS = model.OSFCOS
	m := New(w)

	got := m.maxCursor()
	if got != 2 {
		t.Errorf("maxCursor(StepUpdate, FCOS) = %d, want 2", got)
	}
}

func TestMaxCursor_WelcomeChannelCards_Flatcar(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	w.State.Config.OS = model.OSFlatcar
	m := New(w)
	m.osSubView = false // OS already chosen, now at channel picker

	got := m.maxCursor()
	if got != 4 {
		t.Errorf("maxCursor(StepWelcome, Flatcar channel picker) = %d, want 4", got)
	}
}

func TestMaxCursor_WelcomeChannelCards_FCOS(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	w.State.Config.OS = model.OSFCOS
	m := New(w)
	m.osSubView = false // OS already chosen, now at stream picker

	got := m.maxCursor()
	if got != 3 {
		t.Errorf("maxCursor(StepWelcome, FCOS stream picker) = %d, want 3", got)
	}
}

// ── FCOS update strategy ──────────────────────────────────────────────────────

func TestViewUpdate_FCOS_ShowsZincatiOptions(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepUpdate
	w.State.Config.OS = model.OSFCOS
	m := New(w)

	out := m.viewUpdate()
	if !strings.Contains(out, "Fedora CoreOS") {
		t.Errorf("FCOS viewUpdate should mention Fedora CoreOS: %q", out[:min(300, len(out))])
	}
	if !strings.Contains(out, "immediate") {
		t.Errorf("FCOS viewUpdate should show 'immediate': %q", out[:min(300, len(out))])
	}
	if !strings.Contains(out, "disabled") {
		t.Errorf("FCOS viewUpdate should show 'disabled': %q", out[:min(300, len(out))])
	}
	// Flatcar options should not appear
	if strings.Contains(out, "etcd-lock") {
		t.Errorf("FCOS viewUpdate should not show 'etcd-lock': %q", out[:min(300, len(out))])
	}
}

func TestHandleEnter_Update_FCOS_AppliesStrategy(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepUpdate
	w.State.Config.OS = model.OSFCOS

	m := New(w)
	m.cursor = 0 // immediate
	_, _ = m.handleEnter()

	if m.Wizard.State.Config.UpdateStrategy.FCOSUpdateStrategy != model.FCOSStrategyImmediate {
		t.Errorf("expected FCOSUpdateStrategy=%q, got %q",
			model.FCOSStrategyImmediate, m.Wizard.State.Config.UpdateStrategy.FCOSUpdateStrategy)
	}
}

func TestHandleEnter_Update_FCOS_DisabledStrategy(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepUpdate
	w.State.Config.OS = model.OSFCOS

	m := New(w)
	m.cursor = 1 // disabled
	_, _ = m.handleEnter()

	if m.Wizard.State.Config.UpdateStrategy.FCOSUpdateStrategy != model.FCOSStrategyDisabled {
		t.Errorf("expected FCOSUpdateStrategy=%q, got %q",
			model.FCOSStrategyDisabled, m.Wizard.State.Config.UpdateStrategy.FCOSUpdateStrategy)
	}
}

// ── FCOS hostname default ─────────────────────────────────────────────────────

func TestInitForm_FCOSHostnameDefault(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepUser
	w.State.Config.OS = model.OSFCOS

	m := New(w)
	m.initForm()

	if m.Wizard.State.Config.Hostname != "fcos" {
		t.Errorf("FCOS hostname default should be 'fcos', got %q", m.Wizard.State.Config.Hostname)
	}
}

func TestInitForm_FlatcarHostnameDefault(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepUser
	w.State.Config.OS = model.OSFlatcar

	m := New(w)
	m.initForm()

	if m.Wizard.State.Config.Hostname != "flatcar" {
		t.Errorf("Flatcar hostname default should be 'flatcar', got %q", m.Wizard.State.Config.Hostname)
	}
}

// ── FCOS review summary ───────────────────────────────────────────────────────

func TestReviewSummary_FCOSShowsOSRow(t *testing.T) {
	w := newTestWizard()
	w.State.Config.OS = model.OSFCOS
	w.State.Config.Channel = "stable"
	w.State.Config.Hostname = "fcos-node"
	m := New(w)

	out := m.reviewSummary()
	if !strings.Contains(out, "OS: fcos") {
		t.Errorf("reviewSummary should show OS row for FCOS: %q", out)
	}
}

func TestBuildReviewForm_FCOSTitle(t *testing.T) {
	w := newTestWizard()
	w.State.Config.OS = model.OSFCOS
	w.State.Config.Channel = "stable"
	w.State.Config.Hostname = "fcos"
	m := New(w)

	// The title is rendered inside the huh form — exercise the build path
	form := m.buildReviewForm()
	if form == nil {
		t.Error("buildReviewForm() returned nil for FCOS")
	}
}

// ── FCOS viewInstall / viewDone ───────────────────────────────────────────────

func TestViewInstall_FCOS(t *testing.T) {
	w := newTestWizard()
	w.State.Config.OS = model.OSFCOS
	m := New(w)

	out := m.viewInstall()
	if !strings.Contains(out, "Fedora CoreOS") {
		t.Errorf("viewInstall should say 'Fedora CoreOS' for FCOS: %q", out)
	}
}

func TestViewDone_FCOS_Links(t *testing.T) {
	w := newTestWizard()
	w.State.Config.OS = model.OSFCOS
	w.State.Config.Channel = "stable"
	m := New(w)

	out := m.viewDone()
	if !strings.Contains(out, "Fedora CoreOS") {
		t.Errorf("viewDone should mention 'Fedora CoreOS' for FCOS: %q", out[:min(300, len(out))])
	}
	if !strings.Contains(out, "fedoraproject.org") {
		t.Errorf("viewDone should show Fedora community link: %q", out[:min(300, len(out))])
	}
	// Flatcar-specific text should not appear
	if strings.Contains(out, "flatcar.org") {
		t.Errorf("viewDone should not show flatcar.org for FCOS: %q", out[:min(300, len(out))])
	}
}
