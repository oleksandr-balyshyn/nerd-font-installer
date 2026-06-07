package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/w0rxbend/nerd-font-installer/internal/nerdfonts"
)

func TestModelSelectsReleaseAndFamilies(t *testing.T) {
	m := newModel([]nerdfonts.Release{
		{
			Name:     "v3.4.0",
			TagName:  "v3.4.0",
			Families: []string{"Hack", "JetBrainsMono"},
		},
	}, "/tmp/fonts", true, IconAuto)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = requireModel(t, next)
	if m.step != stepFamilies {
		t.Fatalf("step = %v, want stepFamilies", m.step)
	}
	if m.selectedRelease.TagName != "v3.4.0" {
		t.Fatalf("selectedRelease = %q", m.selectedRelease.TagName)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = requireModel(t, next)
	if m.selectedCount() != 1 {
		t.Fatalf("selectedCount() = %d, want 1", m.selectedCount())
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = requireModel(t, next)
	if m.selectedCount() != 2 {
		t.Fatalf("selectedCount() = %d, want 2", m.selectedCount())
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = requireModel(t, next)
	if m.selectedCount() != 0 {
		t.Fatalf("selectedCount() = %d, want 0", m.selectedCount())
	}
}

func TestFamilyItemsUseNerdFontCheckboxesAndIcons(t *testing.T) {
	m := newModel([]nerdfonts.Release{
		{
			Name:     "v3.4.0",
			TagName:  "v3.4.0",
			Families: []string{"Hack", "JetBrainsMono"},
		},
	}, "/tmp/fonts", true, IconNerd)
	m.selectedRelease = m.releases[0]
	m.selectedFamilies = map[string]bool{"Hack": true}

	items := m.familyItems()
	hack := requireItem(t, items[0])
	jetBrains := requireItem(t, items[1])

	if !strings.Contains(hack.title, "󰄲") {
		t.Fatalf("Hack title = %q, want checked box", hack.title)
	}
	if !strings.Contains(hack.title, "󰌌") {
		t.Fatalf("Hack title = %q, want Hack icon", hack.title)
	}
	if !strings.Contains(jetBrains.title, "󰄱") {
		t.Fatalf("JetBrainsMono title = %q, want unchecked box", jetBrains.title)
	}
	if !strings.Contains(jetBrains.title, "") {
		t.Fatalf("JetBrainsMono title = %q, want JetBrains icon", jetBrains.title)
	}
}

func TestFamilyItemsDefaultToUnicodeIcons(t *testing.T) {
	m := newModel([]nerdfonts.Release{
		{
			Name:     "v3.4.0",
			TagName:  "v3.4.0",
			Families: []string{"Hack"},
		},
	}, "/tmp/fonts", true, IconAuto)
	m.selectedRelease = m.releases[0]

	items := m.familyItems()
	hack := requireItem(t, items[0])

	if !strings.Contains(hack.title, "☐") {
		t.Fatalf("Hack title = %q, want unicode checkbox", hack.title)
	}
	if strings.Contains(hack.title, "󰌌") {
		t.Fatalf("Hack title = %q, should not require Nerd Font glyphs by default", hack.title)
	}
}

func TestModelHandlesWindowSizeBeforeFamilyListExists(t *testing.T) {
	m := newModel([]nerdfonts.Release{
		{
			Name:     "v3.4.0",
			TagName:  "v3.4.0",
			Families: []string{"Hack"},
		},
	}, "/tmp/fonts", true, IconAuto)

	next, _ := m.Update(tea.WindowSizeMsg{Width: 96, Height: 12})
	if _, ok := next.(model); !ok {
		t.Fatalf("Update() = %T, want model", next)
	}
}

func requireModel(t *testing.T, got tea.Model) model {
	t.Helper()

	m, ok := got.(model)
	if !ok {
		t.Fatalf("model = %T, want tui.model", got)
	}
	return m
}

func requireItem(t *testing.T, got list.Item) item {
	t.Helper()

	i, ok := got.(item)
	if !ok {
		t.Fatalf("item = %T, want tui.item", got)
	}
	return i
}
