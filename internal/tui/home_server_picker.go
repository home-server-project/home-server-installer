// Modified by Home Server Project from Project Bluefin Knuckle.
package tui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/projectbluefin/knuckle/internal/model"
)

const (
	homeServerFamilyGina      = "family:gina"
	homeServerFamilyUCore     = "family:ucore"
	homeServerFamilyNvidia    = "family:ucore-nvidia"
	homeServerFamilyNvidiaLTS = "family:ucore-nvidia-lts"
)

type homeServerImageFamily struct {
	id   string
	name string
	desc string
}

type homeServerEditionOption struct {
	id   string
	name string
	desc string
}

var homeServerImagesByFamily = map[string][]homeServerEditionOption{
	homeServerFamilyGina: {
		{model.HomeServerUCoreImage, "Home Server Gina LTS", "Home Server Project image based on Universal Blue uCore LTS."},
		{model.HomeServerUCoreHCIImage, "Home Server Gina HCI LTS", "Home Server Project image based on uCore HCI LTS for virtualization hosts."},
	},
	homeServerFamilyUCore: {
		{model.UpstreamUCoreMinimalImage, "uCore Minimal LTS", "Upstream Universal Blue lightweight image."},
		{model.UpstreamUCoreImage, "uCore LTS", "Upstream Universal Blue server image."},
		{model.UpstreamUCoreHCIImage, "uCore HCI LTS", "Upstream Universal Blue HCI image with virtualization tooling."},
	},
	homeServerFamilyNvidia: {
		{model.UpstreamUCoreMinimalNvidiaImage, "uCore Minimal LTS NVIDIA Open", "Lightweight uCore image with the NVIDIA Open driver."},
		{model.UpstreamUCoreNvidiaImage, "uCore LTS NVIDIA Open", "uCore server image with the NVIDIA Open driver."},
		{model.UpstreamUCoreHCINvidiaImage, "uCore HCI LTS NVIDIA Open", "uCore HCI image with the NVIDIA Open driver."},
	},
	homeServerFamilyNvidiaLTS: {
		{model.UpstreamUCoreMinimalNvidiaLTSImage, "uCore Minimal LTS NVIDIA LTS", "Lightweight uCore image with the NVIDIA LTS driver."},
		{model.UpstreamUCoreNvidiaLTSImage, "uCore LTS NVIDIA LTS", "uCore server image with the NVIDIA LTS driver."},
		{model.UpstreamUCoreHCINvidiaLTSImage, "uCore HCI LTS NVIDIA LTS", "uCore HCI image with the NVIDIA LTS driver."},
	},
}

func homeServerOptionsForFamily(family string) []homeServerEditionOption {
	return homeServerImagesByFamily[family]
}

func homeServerFamilyForImage(image string) (string, int, bool) {
	for _, family := range homeServerImageOptions {
		for i, opt := range homeServerOptionsForFamily(family.id) {
			if opt.id == image {
				return family.id, i, true
			}
		}
	}
	return "", 0, false
}

func homeServerImageDisplayName(image string) (string, bool) {
	for _, family := range homeServerImageOptions {
		for _, opt := range homeServerOptionsForFamily(family.id) {
			if opt.id == image {
				return opt.name, true
			}
		}
	}
	return "", false
}

func homeServerFamilyDisplayName(family string) string {
	for _, opt := range homeServerImageOptions {
		if opt.id == family {
			return strings.ReplaceAll(opt.name, "\n  ", " / ")
		}
	}
	return ""
}

func homeServerFamilyIndex(family string) int {
	for i, opt := range homeServerImageOptions {
		if opt.id == family {
			return i
		}
	}
	return 0
}

func renderHomeServerCards(title string, cursor int, namesAndDescriptions [][2]string) string {
	var b strings.Builder

	selectedBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")).
		Padding(0, 1).
		Width(60)
	normalBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(60)
	nameSelected := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	nameNormal := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	b.WriteString("  " + title + "\n\n")
	for i, item := range namesAndDescriptions {
		selected := i == cursor
		prefix := "  "
		nameStyle := nameNormal
		if selected {
			prefix = cursorStyle.Render("▸ ")
			nameStyle = nameSelected
		}

		var card strings.Builder
		card.WriteString(prefix + nameStyle.Render(item[0]) + "\n")
		card.WriteString("  " + descStyle.Render(item[1]))

		if selected {
			b.WriteString(selectedBorder.Render(card.String()))
		} else {
			b.WriteString(normalBorder.Render(card.String()))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dim.Render("  All installer choices use LTS. Switch to stable/testing later with bootc."))
	b.WriteString("\n")
	b.WriteString(dim.Render("  ↑↓/jk select · enter continue"))
	b.WriteString("\n")
	return b.String()
}

func (m *Model) viewHomeServerPicker() string {
	if m.homeServerFamily == "" {
		items := make([][2]string, 0, len(homeServerImageOptions))
		for _, family := range homeServerImageOptions {
			items = append(items, [2]string{family.name, family.desc})
		}
		return renderHomeServerCards("Select installation family:", m.cursor, items)
	}

	options := homeServerOptionsForFamily(m.homeServerFamily)
	items := make([][2]string, 0, len(options))
	for _, opt := range options {
		items = append(items, [2]string{opt.name, opt.desc})
	}
	title := "Select edition:"
	if familyName := homeServerFamilyDisplayName(m.homeServerFamily); familyName != "" {
		title = "Select " + familyName + " edition:"
	}
	return renderHomeServerCards(title, m.cursor, items)
}
