package tui

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/projectbluefin/knuckle/internal/model"
)

// buildHomeServerUserForm keeps the normal identity/authentication workflow but
// makes the Home Server disk layout an explicit choice. Both presets are always
// available; current non-NVIDIA images default to 1 GiB.
func (m *Model) buildHomeServerUserForm() *huh.Form {
	cfg := &m.Wizard.State.Config
	if cfg.HomeServerBootSizeMiB == 0 {
		cfg.HomeServerBootSizeMiB = model.HomeServerBootStandardMiB
	}

	bootOptions := []huh.Option[int]{
		huh.NewOption("1 GiB — Standard (recommended for uCore / uCore HCI)", model.HomeServerBootStandardMiB),
		huh.NewOption("2 GiB — Large (recommended for NVIDIA / custom images)", model.HomeServerBootLargeMiB),
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Storage Layout").
				Description("Choose the XBOOTLDR /boot size. EFI is 512 MiB and the remaining disk is XFS root."),
			huh.NewSelect[int]().
				Title("Boot partition size").
				Description("Both layouts are supported. The recommendation is a default, not a restriction.").
				Options(bootOptions...).
				Value(&cfg.HomeServerBootSizeMiB),
		),
		huh.NewGroup(
			huh.NewNote().
				Title("System Identity").
				Description("Configure the hostname and primary user account."),
			huh.NewInput().
				Title("Hostname").
				Placeholder("homeserver").
				Value(&cfg.Hostname).
				Validate(validateOptionalHostname),
			huh.NewInput().
				Title("Timezone").
				Placeholder("UTC").
				Description("e.g. America/New_York, Europe/Berlin").
				Value(&cfg.Timezone).
				Validate(validateTimezoneInput),
			huh.NewInput().
				Title("Username").
				Description("Primary user account").
				Value(&m.usernameInput).
				Validate(validateOptionalUsername),
			huh.NewInput().
				Title("Password").
				Description("Optional — leave blank for key-only auth").
				EchoMode(huh.EchoModePassword).
				Value(&m.passwordInput),
		),
		huh.NewGroup(
			huh.NewNote().
				Title("Authentication").
				Description("Set up SSH access. A public key embedded by the Home Server Builder is included automatically when present. You can also add a GitHub username or paste a public key directly."),
			huh.NewInput().
				Title("GitHub Username").
				Description("Fetches your SSH public keys automatically").
				Placeholder("username or @username").
				Value(&m.githubUserInput),
			huh.NewInput().
				Title("SSH Public Key").
				Description("Or paste key directly (separate multiple with ;)").
				Value(&m.sshKeyInput),
			huh.NewNote().
				Title("").
				Description(m.homeServerKeysSummary()),
		),
	).WithTheme(huh.ThemeFunc(huh.ThemeDracula)).WithShowHelp(true).WithWidth(80)
}

func (m *Model) buildHomeServerReviewForm() *huh.Form {
	cfg := &m.Wizard.State.Config
	title := fmt.Sprintf("⚠️  DESTRUCTIVE OPERATION — Install %s to disk?", installTargetDisplayName(cfg))

	dangerTheme := huh.ThemeFunc(func(isDark bool) *huh.Styles {
		styles := huh.ThemeDracula(isDark)
		styles.Focused.Title = styles.Focused.Title.
			Foreground(lipgloss.Color("196")).
			Bold(true)
		styles.Focused.Description = styles.Focused.Description.
			Foreground(lipgloss.Color("255"))
		styles.Focused.FocusedButton = styles.Focused.FocusedButton.
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("196"))
		styles.Focused.BlurredButton = styles.Focused.BlurredButton.
			Foreground(lipgloss.Color("252"))
		return styles
	})

	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(m.homeServerReviewSummary()).
				Affirmative("Yes — erase disk and install").
				Negative("No — go back (safe)").
				Value(&m.Wizard.State.Confirmed),
		),
	).WithTheme(dangerTheme).WithShowHelp(true).WithWidth(m.reviewFormWidth())
}

func (m *Model) homeServerReviewSummary() string {
	cfg := &m.Wizard.State.Config
	bootMiB := cfg.HomeServerBootSizeMiB
	if bootMiB == 0 {
		bootMiB = model.HomeServerBootStandardMiB
	}

	summary := m.reviewSummary()
	needle := fmt.Sprintf("  Destination: %s\n", cfg.HomeServerImage)
	replacement := fmt.Sprintf("  Destination: %s\n  /boot: %d GiB XBOOTLDR (ext4)\n  EFI: 512 MiB\n  Root: remaining disk (XFS)\n", cfg.HomeServerImage, bootMiB/1024)
	if strings.Contains(summary, needle) {
		return strings.Replace(summary, needle, replacement, 1)
	}
	return summary
}
