package sdk_i18n

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

//go:embed locales/*.yaml
var sdkLocalesFS embed.FS

const DefaultLocale = "en"

var i18nTokenRegexp = regexp.MustCompile(`t\(([^)]+)\)`)

// flatMap is a map of dot-separated keys to translated strings.
type flatMap map[string]string

// Translator holds translations loaded from SDK embedded files and project files.
// projectLayers is ordered by priority: index 0 has the highest priority.
type Translator struct {
	mu              sync.RWMutex
	projectLayers   []map[string]flatMap
	sdkTranslations map[string]flatMap
}

// projectLocalesPath is the conventional path for project-level locale files.
const projectLocalesPath = "./locales"

// NewTranslator creates a Translator loading, in priority order:
//  1. Each extra path passed by the caller (index 0 = highest priority)
//  2. The project's own ./locales directory (always included)
//  3. SDK embedded translations
//
// Empty strings in paths are ignored. Non-existent directories are silently skipped.
func NewTranslator(paths ...string) *Translator {
	t := &Translator{
		projectLayers:   []map[string]flatMap{},
		sdkTranslations: make(map[string]flatMap),
	}

	// Load SDK embedded translations.
	entries, err := sdkLocalesFS.ReadDir("locales")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
				continue
			}
			locale := strings.TrimSuffix(entry.Name(), ".yaml")
			data, err := sdkLocalesFS.ReadFile("locales/" + entry.Name())
			if err != nil {
				continue
			}
			flat, err := parseYAML(data)
			if err != nil {
				continue
			}
			t.sdkTranslations[locale] = flat
		}
	}

	// Load caller-supplied paths first (highest priority), then the project locales.
	allPaths := append(paths, projectLocalesPath)
	for _, p := range allPaths {
		if p == "" {
			continue
		}
		layer := loadLocalesFromPath(p)
		if len(layer) > 0 {
			t.projectLayers = append(t.projectLayers, layer)
		}
	}

	return t
}

// loadLocalesFromPath reads all *.yaml files in dirPath and returns a locale → flatMap mapping.
// Non-existent directories are silently ignored.
func loadLocalesFromPath(dirPath string) map[string]flatMap {
	layer := make(map[string]flatMap)
	files, err := filepath.Glob(filepath.Join(dirPath, "*.yaml"))
	if err != nil {
		return layer
	}
	for _, file := range files {
		locale := strings.TrimSuffix(filepath.Base(file), ".yaml")
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		flat, err := parseYAML(data)
		if err != nil {
			continue
		}
		layer[locale] = flat
	}
	return layer
}

// NormalizeLocale extracts the primary language tag from an Accept-Language header value.
// Examples:
//
//	"en-US,en;q=0.9,it;q=0.8" → "en"
//	"it-IT"                    → "it"
//	""                         → "en"
func NormalizeLocale(acceptLanguage string) string {
	if acceptLanguage == "" {
		return DefaultLocale
	}
	// Take the first tag (highest priority).
	first := strings.SplitN(acceptLanguage, ",", 2)[0]
	// Strip quality value (;q=...).
	first = strings.SplitN(first, ";", 2)[0]
	// Take only the primary language subtag (before "-").
	lang := strings.SplitN(strings.TrimSpace(first), "-", 2)[0]
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		return DefaultLocale
	}
	return lang
}

// T translates key for the given locale and performs named placeholder interpolation
// using the provided args map. Placeholders in the translated string must use the
// syntax {{key}} (e.g. "Hello {{name}}"). Fallback chain (per locale, then DefaultLocale):
// project layers in priority order → SDK embedded → key itself.
// If args is nil or empty, no interpolation is performed.
func (t *Translator) T(locale, key string, args map[string]any) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	locales := []string{locale}
	if locale != DefaultLocale {
		locales = append(locales, DefaultLocale)
	}

	for _, loc := range locales {
		for _, layer := range t.projectLayers {
			if m, ok := layer[loc]; ok {
				if val, ok := m[key]; ok {
					return interpolateMap(val, args)
				}
			}
		}
		if m, ok := t.sdkTranslations[loc]; ok {
			if val, ok := m[key]; ok {
				return interpolateMap(val, args)
			}
		}
	}

	return key
}

// ResolveTExpr resolves t(<token>) expressions in value using the given locale.
func (t *Translator) ResolveTExpr(locale, value string) string {
	return i18nTokenRegexp.ReplaceAllStringFunc(value, func(match string) string {
		key := match[2 : len(match)-1] // strip leading "t(" and trailing ")"
		return t.T(locale, key, nil)
	})
}

// interpolateMap replaces {{key}} placeholders in val with the corresponding values from args.
func interpolateMap(val string, args map[string]any) string {
	if len(args) == 0 {
		return val
	}
	result := val
	for k, v := range args {
		result = strings.ReplaceAll(result, "{{{"+k+"}}}", fmt.Sprintf("%v", v))
		result = strings.ReplaceAll(result, "{{"+k+"}}", fmt.Sprintf("%v", v))
	}
	return result
}

// parseYAML parses YAML bytes and flattens nested keys using dot notation.
// Example: errors.not_found: "Resource not found"
func parseYAML(data []byte) (flatMap, error) {
	var raw map[interface{}]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	result := make(flatMap)
	flattenMap("", raw, result)
	return result, nil
}

// flattenMap recursively flattens a nested YAML map into dot-separated keys.
func flattenMap(prefix string, m map[interface{}]interface{}, result flatMap) {
	for k, v := range m {
		key := fmt.Sprintf("%v", k)
		if prefix != "" {
			key = prefix + "." + key
		}
		switch val := v.(type) {
		case map[interface{}]interface{}:
			flattenMap(key, val, result)
		case string:
			result[key] = val
		default:
			result[key] = fmt.Sprintf("%v", val)
		}
	}
}
