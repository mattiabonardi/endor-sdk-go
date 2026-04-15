package sdk_i18n_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_i18n"
)

func TestNormalizeLocale(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"en-US,en;q=0.9,it;q=0.8", "en"},
		{"it-IT", "it"},
		{"fr;q=0.9", "fr"},
		{"", "en"},
		{"   ", "en"},
	}
	for _, c := range cases {
		got := sdk_i18n.NormalizeLocale(c.input)
		if got != c.expected {
			t.Errorf("NormalizeLocale(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}

func TestT_SDKEmbeddedFallback(t *testing.T) {
	// Init with empty project path so only SDK embedded translations are used.
	if err := sdk_i18n.Init(""); err != nil {
		t.Fatal(err)
	}
	got := sdk_i18n.T("en", "entities.entity.not_found", nil)
	if got == "entities.entity.not_found" {
		t.Error("expected SDK embedded translation for entities.entity.not_found, got the key itself")
	}
}

func TestT_LocaleFallbackToDefault(t *testing.T) {
	if err := sdk_i18n.Init(""); err != nil {
		t.Fatal(err)
	}
	// "it" locale without project files falls back to "en".
	got := sdk_i18n.T("it", "entities.entity.not_found", nil)
	if got == "entities.entity.not_found" {
		t.Error("expected fallback to SDK en translation, got the key itself")
	}
}

func TestT_ProjectOverridesSDK(t *testing.T) {
	// Write a temporary project locale file that overrides the SDK translation.
	dir := t.TempDir()
	content := "entities:\n  entity:\n    not_found: \"Risorsa non trovata\"\n"
	if err := os.WriteFile(filepath.Join(dir, "it.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := sdk_i18n.Init(dir); err != nil {
		t.Fatal(err)
	}

	got := sdk_i18n.T("it", "entities.entity.not_found", nil)
	if got != "Risorsa non trovata" {
		t.Errorf("expected project translation override, got %q", got)
	}
}

func TestT_Interpolation(t *testing.T) {
	if err := sdk_i18n.Init(""); err != nil {
		t.Fatal(err)
	}
	got := sdk_i18n.T("en", "entities.entity.not_found", map[string]any{"id": "123"})
	if got == "entities.entity.not_found" {
		t.Error("expected interpolated translation, got key itself")
	}
	if got == "entity {{id}} not found" {
		t.Errorf("expected placeholder to be replaced, got %q", got)
	}
}

func TestT_UnknownKeyReturnsKey(t *testing.T) {
	if err := sdk_i18n.Init(""); err != nil {
		t.Fatal(err)
	}
	got := sdk_i18n.T("en", "this.key.does.not.exist", nil)
	if got != "this.key.does.not.exist" {
		t.Errorf("expected key itself for unknown key, got %q", got)
	}
}
