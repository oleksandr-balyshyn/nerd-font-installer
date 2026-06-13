package fontname

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAcceptsSafeNames(t *testing.T) {
	t.Parallel()
	for _, family := range []string{"Hack", "JetBrainsMono", "Symbols Nerd Font", "0xProto", "Fira-Code_1"} {
		if err := Validate(family); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", family, err)
		}
	}
}

func TestValidateRejectsUnsafeNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		family string
	}{
		{name: "empty", family: ""},
		{name: "dot", family: "."},
		{name: "dot dot", family: ".."},
		{name: "slash", family: "Hack/Regular"},
		{name: "backslash", family: `Hack\Regular`},
		{name: "absolute", family: "/tmp/Hack"},
		{name: "traversal", family: "../Hack"},
		{name: "nul byte", family: "Hack\x00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := Validate(tt.family); err == nil {
				t.Fatalf("Validate(%q) = nil, want error", tt.family)
			}
		})
	}
}

// FuzzValidate asserts the security invariant: any name Validate accepts must be
// a benign single path component. This guards the path-traversal boundary that
// both config and fonts rely on.
func FuzzValidate(f *testing.F) {
	for _, seed := range []string{"Hack", "", ".", "..", "../x", `..\x`, "/abs", "a/b", "x\x00y", "Symbols Nerd Font"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, family string) {
		if Validate(family) != nil {
			return
		}
		if family == "" {
			t.Fatalf("accepted empty family name")
		}
		if filepath.IsAbs(family) {
			t.Fatalf("accepted absolute family name %q", family)
		}
		if filepath.Base(family) != family {
			t.Fatalf("accepted family %q whose Base is %q", family, filepath.Base(family))
		}
		if strings.ContainsAny(family, `/\`) {
			t.Fatalf("accepted family %q containing a separator", family)
		}
		if strings.ContainsRune(family, 0) {
			t.Fatalf("accepted family %q containing a NUL byte", family)
		}
	})
}
