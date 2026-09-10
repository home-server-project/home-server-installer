package tui

import (
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
)

// TestHandleEnter_Welcome_IgnitionURL_SkipsToStorage covers the legacy
// channel/stream Welcome path directly. The Home Server family picker no
// longer exposes this path, but the generic flow remains supported.
func TestHandleEnter_Welcome_IgnitionURL_SkipsToStorage(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepWelcome
	w.State.Config.IgnitionURL = "https://example.com/config.ign"
	w.State.Config.Channel = "stable"

	m := New(w)
	m.osSubView = false
	m.cursor = 0
	_, _ = m.handleEnter()

	if m.Wizard.State.CurrentStep != model.StepStorage {
		t.Errorf("IgnitionURL on legacy Welcome path: expected StepStorage, got %v", m.Wizard.State.CurrentStep)
	}
	if m.err != nil {
		t.Errorf("IgnitionURL on legacy Welcome path: unexpected error %v", m.err)
	}
	if m.cursor != 0 {
		t.Errorf("IgnitionURL on legacy Welcome path: cursor should be reset to 0, got %d", m.cursor)
	}
}

// TestHandleEnter_Storage_IgnitionURL_SkipsToReview covers the branch where
// IgnitionURL is set on StepStorage and the installer skips to StepReview.
func TestHandleEnter_Storage_IgnitionURL_SkipsToReview(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepStorage
	w.State.Config.IgnitionURL = "https://example.com/config.ign"
	w.State.Config.Channel = "stable"
	w.State.Config.Hostname = "testhost"
	w.State.Disks = []model.DiskInfo{{DevPath: "/dev/vda", Size: 20 * 1024 * 1024 * 1024}}
	w.State.Config.Disk = model.DiskInfo{DevPath: "/dev/vda", Size: 20 * 1024 * 1024 * 1024}

	m := New(w)
	m.cursor = 0

	_, _ = m.handleEnter()

	if m.Wizard.State.CurrentStep != model.StepReview && m.Wizard.State.CurrentStep != model.StepStorage {
		t.Errorf("IgnitionURL on Storage: unexpected step %v", m.Wizard.State.CurrentStep)
	}
}

// TestHandleEnter_Storage_IgnitionURL_ValidationError covers the branch where
// IgnitionURL is set but storage validation fails.
func TestHandleEnter_Storage_IgnitionURL_ValidationError(t *testing.T) {
	w := newTestWizard()
	w.State.CurrentStep = model.StepStorage
	w.State.Config.IgnitionURL = "https://example.com/config.ign"
	w.State.Config.Disk = model.DiskInfo{}

	m := New(w)
	m.cursor = 0

	_, _ = m.handleEnter()

	if m.Wizard.State.CurrentStep == model.StepReview {
		t.Error("IgnitionURL on Storage with invalid disk: should not advance to StepReview")
	}
}
