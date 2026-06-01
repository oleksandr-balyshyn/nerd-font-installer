package nerdfonts

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestFamiliesFromAssets(t *testing.T) {
	got := familiesFromAssets([]string{
		"JetBrainsMono.zip",
		"README.md",
		"Hack.ZIP",
		"JetBrainsMono.zip",
		"SymbolsOnly.tar.xz",
	})
	want := []string{"Hack", "JetBrainsMono"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("familiesFromAssets() = %#v, want %#v", got, want)
	}
}

func TestWithPage(t *testing.T) {
	got, err := withPage("https://example.test/releases?existing=1", 3)
	if err != nil {
		t.Fatalf("withPage() error = %v", err)
	}
	want := "https://example.test/releases?existing=1&page=3&per_page=100"
	if got != want {
		t.Fatalf("withPage() = %q, want %q", got, want)
	}
}

func TestClientReleasesFetchesAndFiltersPages(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.RawQuery)
		if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
			t.Fatalf("Accept = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "nerdfont-install" {
			t.Fatalf("User-Agent = %q", got)
		}

		switch r.URL.Query().Get("page") {
		case "1":
			writeJSON(t, w, []map[string]any{
				{
					"name":     "",
					"tag_name": "v3.4.0",
					"assets": []map[string]any{
						{"name": "Hack.zip"},
						{"name": "README.md"},
					},
				},
				{
					"name":     "draft",
					"tag_name": "v3.5.0",
					"draft":    true,
					"assets": []map[string]any{
						{"name": "Ignored.zip"},
					},
				},
			})
		case "2":
			writeJSON(t, w, []map[string]any{
				{
					"name":     "No assets",
					"tag_name": "v3.3.0",
					"assets":   []map[string]any{},
				},
				{
					"name":     "v3.2.0",
					"tag_name": "v3.2.0",
					"assets": []map[string]any{
						{"name": "JetBrainsMono.zip"},
					},
				},
			})
		default:
			writeJSON(t, w, []map[string]any{})
		}
	}))
	defer server.Close()

	releases, err := Client{BaseURL: server.URL, MaxPages: 3}.Releases(t.Context())
	if err != nil {
		t.Fatalf("Releases() error = %v", err)
	}

	want := []Release{
		{Name: "v3.4.0", TagName: "v3.4.0", Families: []string{"Hack"}},
		{Name: "v3.2.0", TagName: "v3.2.0", Families: []string{"JetBrainsMono"}},
	}
	if !reflect.DeepEqual(releases, want) {
		t.Fatalf("Releases() = %#v, want %#v", releases, want)
	}
	if len(requests) != 3 {
		t.Fatalf("requests = %#v, want 3 pages", requests)
	}
}

func TestClientReleasesStopsAtMaxPages(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		writeJSON(t, w, []map[string]any{
			{
				"name":     "v3.4.0",
				"tag_name": "v3.4.0",
				"assets": []map[string]any{
					{"name": "Hack.zip"},
				},
			},
		})
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL, MaxPages: 2}
	if _, err := client.Releases(t.Context()); err != nil {
		t.Fatalf("Releases() error = %v", err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}

func TestClientReleasesErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "non 2xx", status: http.StatusForbidden, body: `{"message":"rate limited"}`},
		{name: "malformed json", status: http.StatusOK, body: `[`},
		{name: "empty filtered result", status: http.StatusOK, body: `[]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				if _, err := w.Write([]byte(tt.body)); err != nil {
					t.Fatal(err)
				}
			}))
			defer server.Close()

			_, err := Client{BaseURL: server.URL, MaxPages: 1}.Releases(t.Context())
			if err == nil {
				t.Fatal("Releases() error = nil, want error")
			}
		})
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}
