package sdk_i18n

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

//go:embed locales/*.yaml
var sdkLocalesFS embed.FS

const DefaultLocale = "en"

// flatMap is a map of dot-separated keys to translated strings.
type flatMap map[string]string

// Translator holds translations loaded from SDK embedded files and project files.
type Translator struct {
	mu                  sync.RWMutex
	projectTranslations map[string]flatMap
	sdkTranslations     map[string]flatMap
}

var globalTranslator *Translator

// Init loads SDK embedded translations and project translations from projectLocalePath.
// projectLocalePath is the path to the project's locale directory (e.g. "./locales").
// It is safe to call with an empty or non-existent path; SDK defaults will still be available.
func Init(projectLocalePath string) error {
	t := &Translator{
		projectTranslations: make(map[string]flatMap),
		sdkTranslations:     make(map[string]flatMap),
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

	// Load project translations from the filesystem.
	if projectLocalePath != "" {
		files, err := filepath.Glob(filepath.Join(projectLocalePath, "*.yaml"))
		if err == nil {
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
				t.projectTranslations[locale] = flat
			}
		}
	}

	globalTranslator = t
	return nil
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
// syntax {{key}} (e.g. "Hello {{name}}"). Fallback chain: project locale → SDK locale →
// project default → SDK default → key itself.
// If args is nil or empty, no interpolation is performed.
func T(locale, key string, args map[string]any) string {
	if globalTranslator == nil {
		return key
	}
	globalTranslator.mu.RLock()
	defer globalTranslator.mu.RUnlock()

	locales := []string{locale}
	if locale != DefaultLocale {
		locales = append(locales, DefaultLocale)
	}

	for _, loc := range locales {
		if m, ok := globalTranslator.projectTranslations[loc]; ok {
			if val, ok := m[key]; ok {
				return interpolateMap(val, args)
			}
		}
		if m, ok := globalTranslator.sdkTranslations[loc]; ok {
			if val, ok := m[key]; ok {
				return interpolateMap(val, args)
			}
		}
	}

	return key
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
