// Package fontname validates Nerd Font family names before they are used as
// single filesystem path components and URL path segments.
//
// The check is security-relevant: a family name is joined onto the destination
// directory (filepath.Join(root, family)) and onto the download URL, so a name
// that escapes its directory via traversal, an absolute path, or a path
// separator must be rejected before it reaches the filesystem or the network.
// Both the config loader and the font installer validate at their own trust
// boundaries; they share this single implementation so the guard cannot drift
// out of agreement between them.
package fontname

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Validate reports an error if family is not safe to use as a single path
// component. A nil result guarantees that family is non-empty, contains no path
// separators or NUL bytes, is not absolute, and equals its own filepath.Base.
func Validate(family string) error {
	switch {
	case family == "":
		return fmt.Errorf("font family names cannot be empty")
	case family == "." || family == "..":
		return fmt.Errorf("unsafe font family name %q", family)
	case strings.ContainsAny(family, `/\`):
		return fmt.Errorf("unsafe font family name %q", family)
	case strings.ContainsRune(family, 0):
		return fmt.Errorf("unsafe font family name %q", family)
	case filepath.IsAbs(family):
		return fmt.Errorf("unsafe font family name %q", family)
	case filepath.Base(family) != family:
		return fmt.Errorf("unsafe font family name %q", family)
	default:
		return nil
	}
}
