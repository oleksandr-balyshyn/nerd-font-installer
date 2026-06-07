package tui

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/w0rxbend/nerd-font-installer/internal/config"
	"github.com/w0rxbend/nerd-font-installer/internal/nerdfonts"
)

type Result struct {
	Config    config.Config
	Cancelled bool
}

type Options struct {
	Destination      string
	RefreshFontCache bool
	Output           io.Writer
	Icons            IconMode
}

type step int

const (
	stepRelease step = iota
	stepFamilies
	stepDone
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63"))
	bannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("81")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 3)
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")).
			Bold(true)
	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("219"))
	pillStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("63")).
			Bold(true).
			Padding(0, 1)
	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63"))
)

const (
	IconAuto    IconMode = "auto"
	IconNerd    IconMode = "nerd"
	IconUnicode IconMode = "unicode"
	IconASCII   IconMode = "ascii"
)

type IconMode string

type iconSet struct {
	Mode       IconMode
	Title      string
	Package    string
	Release    string
	Font       string
	Folder     string
	Checked    string
	Unchecked  string
	Selected   string
	Ready      string
	Launch     string
	Toolbox    string
	Separator  string
	NerdFamily map[string]string
}

type item struct {
	title       string
	description string
	value       string
}

func (i item) Title() string {
	return i.title
}

func (i item) Description() string {
	return i.description
}

func (i item) FilterValue() string {
	return strings.Join([]string{i.title, i.description, i.value}, " ")
}

type model struct {
	step             step
	releases         []nerdfonts.Release
	releaseList      list.Model
	familyList       list.Model
	icons            iconSet
	selectedFamilies map[string]bool
	selectedRelease  nerdfonts.Release
	destination      string
	refreshFontCache bool
	cancelled        bool
	err              error
}

type loadReleasesMsg struct {
	releases []nerdfonts.Release
	err      error
}

type loadingModel struct {
	spinner spinner.Model
	load    func(context.Context) ([]nerdfonts.Release, error)
	ctx     context.Context
	message string
	state   *loadingState
}

type loadingState struct {
	releases []nerdfonts.Release
	err      error
	done     bool
}

func LoadReleases(
	ctx context.Context,
	load func(context.Context) ([]nerdfonts.Release, error),
	output io.Writer,
) ([]nerdfonts.Release, error) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	programOptions := []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithInput(nil),
		tea.WithoutSignalHandler(),
	}
	if output != nil {
		programOptions = append(programOptions, tea.WithOutput(output))
	}

	program := tea.NewProgram(loadingModel{
		spinner: s,
		load:    load,
		ctx:     ctx,
		message: "Loading Nerd Fonts releases",
		state:   &loadingState{},
	}, programOptions...)
	finalModel, err := program.Run()
	if err != nil {
		return nil, err
	}

	m, ok := finalModel.(loadingModel)
	if !ok {
		return nil, fmt.Errorf("unexpected loading model %T", finalModel)
	}
	if !m.state.done {
		return nil, fmt.Errorf("release loader exited before completion")
	}
	return m.state.releases, m.state.err
}

func (m loadingModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		releases, err := m.load(m.ctx)
		return loadReleasesMsg{releases: releases, err: err}
	})
}

func (m loadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadReleasesMsg:
		m.state.releases = msg.releases
		m.state.err = msg.err
		m.state.done = true
		if msg.err != nil {
			m.message = errorStyle.Render(msg.err.Error())
			return m, tea.Quit
		}
		m.message = successStyle.Render("OK  Releases loaded")
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m loadingModel) View() string {
	return fmt.Sprintf("%s %s\n", m.spinner.View(), accentStyle.Render(m.message))
}

func Run(ctx context.Context, releases []nerdfonts.Release, opts Options) (Result, error) {
	if len(releases) == 0 {
		return Result{}, fmt.Errorf("no Nerd Fonts releases available")
	}

	destination := opts.Destination
	if strings.TrimSpace(destination) == "" {
		destination = "~/.local/share/fonts/NerdFonts"
	}

	m := newModel(releases, destination, opts.RefreshFontCache, opts.Icons)
	programOptions := []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithAltScreen(),
	}
	if opts.Output != nil {
		programOptions = append(programOptions, tea.WithOutput(opts.Output))
	}

	program := tea.NewProgram(m, programOptions...)
	finalModel, err := program.Run()
	if err != nil {
		return Result{}, err
	}

	m, ok := finalModel.(model)
	if !ok {
		return Result{}, fmt.Errorf("unexpected TUI model %T", finalModel)
	}
	if m.cancelled {
		return Result{Cancelled: true}, nil
	}
	if m.err != nil {
		return Result{}, m.err
	}

	families := make([]string, 0, len(m.selectedFamilies))
	for family, selected := range m.selectedFamilies {
		if selected {
			families = append(families, family)
		}
	}
	slices.Sort(families)
	if len(families) == 0 {
		return Result{Cancelled: true}, nil
	}

	return Result{
		Config: config.Config{
			Release:          m.selectedRelease.TagName,
			Destination:      m.destination,
			RefreshFontCache: m.refreshFontCache,
			Families:         families,
		},
	}, nil
}

func newModel(releases []nerdfonts.Release, destination string, refreshFontCache bool, iconMode IconMode) model {
	icons := resolveIconSet(iconMode)
	items := make([]list.Item, 0, len(releases))
	for _, release := range releases {
		description := fmt.Sprintf("%s  %d font archives  %s  %s ready for terminals and editors",
			icons.Font,
			len(release.Families),
			icons.Separator,
			icons.Toolbox,
		)
		items = append(items, item{
			title:       icons.Release + " " + release.TagName,
			description: description,
			value:       release.TagName,
		})
	}

	delegate := newDelegate()
	releaseList := list.New(items, delegate, 0, 0)
	releaseList.Title = icons.Package + "  Select Nerd Fonts release"
	releaseList.SetShowStatusBar(false)
	releaseList.SetFilteringEnabled(true)
	releaseList.Styles.Title = releaseList.Styles.Title.
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("63")).
		Bold(true)
	releaseList.Styles.PaginationStyle = helpStyle
	releaseList.Styles.HelpStyle = helpStyle

	return model{
		step:             stepRelease,
		releases:         releases,
		releaseList:      releaseList,
		icons:            icons,
		selectedFamilies: map[string]bool{},
		destination:      destination,
		refreshFontCache: refreshFontCache,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.releaseList = setListSize(m.releaseList, msg.Width, msg.Height-10)
		if m.step == stepFamilies {
			m.familyList = setListSize(m.familyList, msg.Width, msg.Height-12)
		}
		return m, nil
	case tea.KeyMsg:
		return m.updateKey(msg)
	}

	var cmd tea.Cmd
	switch m.step {
	case stepRelease:
		m.releaseList, cmd = m.releaseList.Update(msg)
	case stepFamilies:
		m.familyList, cmd = m.familyList.Update(msg)
	}
	return m, cmd
}

func (m model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.cancelled = true
		return m, tea.Quit
	case "esc":
		if m.step == stepFamilies {
			m.step = stepRelease
			return m, nil
		}
		m.cancelled = true
		return m, tea.Quit
	}

	switch m.step {
	case stepRelease:
		return m.updateReleaseKey(msg)
	case stepFamilies:
		return m.updateFamilyKey(msg)
	default:
		return m, nil
	}
}

func (m model) updateReleaseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() != "enter" {
		var cmd tea.Cmd
		m.releaseList, cmd = m.releaseList.Update(msg)
		return m, cmd
	}

	selected, ok := m.releaseList.SelectedItem().(item)
	if !ok {
		return m, nil
	}
	for _, release := range m.releases {
		if release.TagName == selected.value {
			m.selectedRelease = release
			m.step = stepFamilies
			m.selectedFamilies = map[string]bool{}
			m.familyList = m.newFamilyList()
			return m, nil
		}
	}

	m.err = fmt.Errorf("selected release %q was not found", selected.title)
	return m, tea.Quit
}

func (m model) updateFamilyKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b":
		m.step = stepRelease
		return m, nil
	case " ":
		selected, ok := m.familyList.SelectedItem().(item)
		if !ok {
			return m, nil
		}
		m.selectedFamilies[selected.value] = !m.selectedFamilies[selected.value]
		m.familyList.SetItems(m.familyItems())
		return m, nil
	case "a":
		allSelected := m.selectedCount() == len(m.selectedRelease.Families)
		m.selectedFamilies = map[string]bool{}
		if !allSelected {
			for _, family := range m.selectedRelease.Families {
				m.selectedFamilies[family] = true
			}
		}
		m.familyList.SetItems(m.familyItems())
		return m, nil
	case "enter":
		if m.selectedCount() == 0 {
			return m, nil
		}
		m.step = stepDone
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.familyList, cmd = m.familyList.Update(msg)
		return m, cmd
	}
}

func (m model) newFamilyList() list.Model {
	delegate := newDelegate()
	familyList := list.New(m.familyItems(), delegate, m.releaseList.Width(), m.releaseList.Height())
	familyList.Title = m.icons.Title + "  Select font families"
	familyList.SetShowStatusBar(false)
	familyList.SetFilteringEnabled(true)
	familyList.Styles.Title = familyList.Styles.Title.
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("63")).
		Bold(true)
	familyList.Styles.PaginationStyle = helpStyle
	familyList.Styles.HelpStyle = helpStyle
	return familyList
}

func (m model) familyItems() []list.Item {
	items := make([]list.Item, 0, len(m.selectedRelease.Families))
	for _, family := range m.selectedRelease.Families {
		marker := m.icons.Unchecked
		if m.selectedFamilies[family] {
			marker = m.icons.Checked
		}
		items = append(items, item{
			title:       marker + "  " + m.iconForFamily(family) + "  " + family,
			description: fmt.Sprintf("%s %s  %s  %s", m.icons.Release, m.selectedRelease.TagName, m.icons.Separator, familyHint(family)),
			value:       family,
		})
	}
	return items
}

func newDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("230")).
		BorderForeground(lipgloss.Color("63")).
		Background(lipgloss.Color("57")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("219")).
		BorderForeground(lipgloss.Color("63"))
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(lipgloss.Color("252")).Bold(true)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(lipgloss.Color("245"))
	delegate.Styles.FilterMatch = delegate.Styles.FilterMatch.Foreground(lipgloss.Color("214")).Bold(true)
	return delegate
}

func (m model) iconForFamily(family string) string {
	key := strings.ToLower(strings.ReplaceAll(family, " ", ""))
	if icon, ok := m.icons.NerdFamily[key]; ok {
		return icon
	}
	return m.icons.Font
}

func familyHint(family string) string {
	key := strings.ToLower(family)
	switch {
	case strings.Contains(key, "mono"):
		return "monospace favorite"
	case strings.Contains(key, "code"):
		return "coding ligatures"
	case strings.Contains(key, "symbol"):
		return "glyph toolkit"
	default:
		return "Nerd Font patched"
	}
}

func resolveIconSet(mode IconMode) iconSet {
	switch mode {
	case IconNerd:
		return iconSet{
			Mode:      IconNerd,
			Title:     "󰛖",
			Package:   "",
			Release:   "󰐕",
			Font:      "",
			Folder:    "",
			Checked:   "󰄲",
			Unchecked: "󰄱",
			Selected:  "✅",
			Ready:     "✅",
			Launch:    "🚀",
			Toolbox:   "🧰",
			Separator: "•",
			NerdFamily: map[string]string{
				"0xproto":         "",
				"adwaitamono":     "",
				"anonymouspro":    "󰈙",
				"caskaydiacove":   "",
				"cascadiacode":    "",
				"cascadiamono":    "",
				"firacode":        "",
				"firago":          "",
				"hack":            "󰌌",
				"ibmplexmono":     "󰡱",
				"iosevka":         "󰘦",
				"jetbrainsmono":   "",
				"meslo":           "",
				"monaspace":       "",
				"robotomono":      "󱚤",
				"saucecodepro":    "",
				"spacemono":       "󰎆",
				"symbolsnerdfont": "󰣆",
				"ubuntu":          "",
				"ubuntumono":      "",
				"victormono":      "󰘦",
			},
		}
	case IconASCII:
		return iconSet{
			Mode:       IconASCII,
			Title:      "NF",
			Package:    "pkg",
			Release:    "tag",
			Font:       "Aa",
			Folder:     "dir",
			Checked:    "[x]",
			Unchecked:  "[ ]",
			Selected:   "OK",
			Ready:      "OK",
			Launch:     ">>",
			Toolbox:    "tools",
			Separator:  "-",
			NerdFamily: map[string]string{},
		}
	default:
		return iconSet{
			Mode:       IconUnicode,
			Title:      "✦",
			Package:    "▣",
			Release:    "◆",
			Font:       "Aa",
			Folder:     "⌂",
			Checked:    "☑",
			Unchecked:  "☐",
			Selected:   "✓",
			Ready:      "✓",
			Launch:     "→",
			Toolbox:    "◇",
			Separator:  "•",
			NerdFamily: map[string]string{},
		}
	}
}

func setListSize(model list.Model, width, height int) (resized list.Model) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	defer func() {
		if recover() != nil {
			resized = model
		}
	}()
	model.SetSize(width, height)
	return model
}

func (m model) selectedCount() int {
	count := 0
	for _, selected := range m.selectedFamilies {
		if selected {
			count++
		}
	}
	return count
}

func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(m.err.Error())
	}

	switch m.step {
	case stepRelease:
		return strings.Join([]string{
			m.banner(),
			subtitleStyle.Render(m.icons.Launch + " Choose a release, then collect patched fonts with devicons, symbols, and terminal glyphs."),
			m.releaseList.View(),
			help("enter", "choose release", "/", "filter", "q", "quit"),
		}, "\n")
	case stepFamilies:
		summary := fmt.Sprintf(
			"%s  %s  %s",
			pillStyle.Render(fmt.Sprintf("%s %s", m.icons.Release, m.selectedRelease.TagName)),
			pathStyle.Render(m.icons.Folder+" "+m.destination),
			successStyle.Render(fmt.Sprintf("%s %d selected", m.icons.Selected, m.selectedCount())),
		)
		return strings.Join([]string{
			m.banner(),
			summary,
			m.familyList.View(),
			help("space", "toggle", "a", "all/none", "enter", "install", "b/esc", "back", "/", "filter", "q", "quit"),
		}, "\n")
	case stepDone:
		return successStyle.Render(m.icons.Ready + "  Ready to install selected fonts")
	default:
		return ""
	}
}

func (m model) banner() string {
	lines := []string{
		titleStyle.Render(m.icons.Title + "  Nerd Font Installer  " + m.icons.Package + "  " + m.icons.Font),
		subtitleStyle.Render("Terminal fonts, devicons, ligatures, and patched glyphs in one pass " + m.icons.Launch),
	}
	return bannerStyle.Render(strings.Join(lines, "\n"))
}

func help(parts ...string) string {
	if len(parts)%2 != 0 {
		return helpStyle.Render(strings.Join(parts, " "))
	}

	segments := make([]string, 0, len(parts)/2)
	for i := 0; i < len(parts); i += 2 {
		segments = append(segments, keyStyle.Render(parts[i])+helpStyle.Render(": "+parts[i+1]))
	}
	return strings.Join(segments, helpStyle.Render("  •  "))
}
