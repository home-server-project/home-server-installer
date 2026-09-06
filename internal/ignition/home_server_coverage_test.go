package ignition

import (
	"strings"
	"testing"

	"github.com/projectbluefin/knuckle/internal/model"
)

func homeServerCoverageConfig() *model.InstallConfig {
	return &model.InstallConfig{
		OS:              model.OSFCOS,
		HomeServerImage: model.HomeServerUCoreImage,
		Channel:         "stable",
		Hostname:        "homeserver",
		Timezone:        "UTC",
		SSHKeys:         []string{"ssh-ed25519 AAAATEST home@test"},
	}
}

func TestGenerateHomeServerFCOSButaneNilConfig(t *testing.T) {
	if _, err := NewGenerator().GenerateHomeServerFCOSButane(nil); err == nil {
		t.Fatal("expected nil config to fail")
	}
}

func TestGenerateHomeServerFCOSButaneSharedFragmentError(t *testing.T) {
	orig := hostnameTemplate
	hostnameTemplate = "{{"
	t.Cleanup(func() { hostnameTemplate = orig })

	if _, err := NewGenerator().GenerateHomeServerFCOSButane(homeServerCoverageConfig()); err == nil {
		t.Fatal("expected shared-fragment rendering error")
	}
}

func TestGenerateHomeServerFCOSButaneKeyTemplateError(t *testing.T) {
	orig := homeServerKeyTemplate
	homeServerKeyTemplate = "{{"
	t.Cleanup(func() { homeServerKeyTemplate = orig })

	_, err := NewGenerator().GenerateHomeServerFCOSButane(homeServerCoverageConfig())
	if err == nil || !strings.Contains(err.Error(), "signing key") {
		t.Fatalf("expected signing-key rendering error, got %v", err)
	}
}

func TestGenerateHomeServerFCOSButanePolicyTemplateError(t *testing.T) {
	orig := homeServerPolicyTemplate
	homeServerPolicyTemplate = "{{"
	t.Cleanup(func() { homeServerPolicyTemplate = orig })

	_, err := NewGenerator().GenerateHomeServerFCOSButane(homeServerCoverageConfig())
	if err == nil || !strings.Contains(err.Error(), "signature policy") {
		t.Fatalf("expected signature-policy rendering error, got %v", err)
	}
}

func TestGenerateHomeServerFCOSButaneAutorebaseScriptTemplateError(t *testing.T) {
	orig := homeServerAutorebaseScriptTemplate
	homeServerAutorebaseScriptTemplate = "{{"
	t.Cleanup(func() { homeServerAutorebaseScriptTemplate = orig })

	_, err := NewGenerator().GenerateHomeServerFCOSButane(homeServerCoverageConfig())
	if err == nil || !strings.Contains(err.Error(), "autorebase script") {
		t.Fatalf("expected autorebase-script rendering error, got %v", err)
	}
}

func TestGenerateHomeServerFCOSButaneCleanupScriptTemplateError(t *testing.T) {
	orig := homeServerPostRebaseCleanupScriptTemplate
	homeServerPostRebaseCleanupScriptTemplate = "{{"
	t.Cleanup(func() { homeServerPostRebaseCleanupScriptTemplate = orig })

	_, err := NewGenerator().GenerateHomeServerFCOSButane(homeServerCoverageConfig())
	if err == nil || !strings.Contains(err.Error(), "post-rebase cleanup script") {
		t.Fatalf("expected cleanup-script rendering error, got %v", err)
	}
}

func TestGenerateHomeServerFCOSButaneServiceTemplateError(t *testing.T) {
	orig := homeServerAutorebaseTemplate
	homeServerAutorebaseTemplate = "{{"
	t.Cleanup(func() { homeServerAutorebaseTemplate = orig })

	_, err := NewGenerator().GenerateHomeServerFCOSButane(homeServerCoverageConfig())
	if err == nil || !strings.Contains(err.Error(), "autorebase service") {
		t.Fatalf("expected autorebase-service rendering error, got %v", err)
	}
}

func TestGenerateHomeServerFCOSButaneCleanupServiceTemplateError(t *testing.T) {
	orig := homeServerPostRebaseCleanupTemplate
	homeServerPostRebaseCleanupTemplate = "{{"
	t.Cleanup(func() { homeServerPostRebaseCleanupTemplate = orig })

	_, err := NewGenerator().GenerateHomeServerFCOSButane(homeServerCoverageConfig())
	if err == nil || !strings.Contains(err.Error(), "post-rebase cleanup service") {
		t.Fatalf("expected cleanup-service rendering error, got %v", err)
	}
}

func TestGenerateHomeServerFCOSButaneHCI(t *testing.T) {
	cfg := homeServerCoverageConfig()
	cfg.HomeServerImage = model.HomeServerUCoreHCIImage
	out, err := NewGenerator().GenerateHomeServerFCOSButane(cfg)
	if err != nil {
		t.Fatalf("GenerateHomeServerFCOSButane HCI: %v", err)
	}
	if !strings.Contains(out, model.HomeServerUCoreHCIImage) {
		t.Fatalf("HCI output missing selected image: %s", out)
	}
}
