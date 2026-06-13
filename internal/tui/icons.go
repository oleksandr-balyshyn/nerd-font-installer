package tui

// This file holds the icon-set data tables. They are reference data, not logic,
// so they live in package-level vars; resolveIconSet is just a lookup. Adding a
// font means editing nerdFamilyGlyphs, not a switch arm.

// nerdFamilyGlyphs maps a normalized family key (lowercased, spaces removed) to
// its Nerd Font glyph, used only when IconNerd is selected.
var nerdFamilyGlyphs = map[string]string{
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
}

var nerdIcons = iconSet{
	Mode:       IconNerd,
	Title:      "󰛖",
	Package:    "",
	Release:    "󰐕",
	Font:       "",
	Folder:     "",
	Checked:    "󰄲",
	Unchecked:  "󰄱",
	Selected:   "✅",
	Ready:      "✅",
	Launch:     "🚀",
	Toolbox:    "🧰",
	Separator:  "•",
	NerdFamily: nerdFamilyGlyphs,
}

var asciiIcons = iconSet{
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

// unicodeIcons is the safe default (also used for IconAuto): expressive glyphs
// that do not require a patched Nerd Font to render.
var unicodeIcons = iconSet{
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

func resolveIconSet(mode IconMode) iconSet {
	switch mode {
	case IconNerd:
		return nerdIcons
	case IconASCII:
		return asciiIcons
	default: // IconAuto and IconUnicode both use the safe Unicode set.
		return unicodeIcons
	}
}
