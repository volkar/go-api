package i18n

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type ctxKey string

const userLangKey ctxKey = "lang"

//go:embed translations.json
var localesFS embed.FS

type Translator struct {
	data         map[string]map[string]string
	fallbackLang string
}

func New(fallbackLang string, logger *slog.Logger) *Translator {
	file, err := localesFS.ReadFile("translations.json")
	if err != nil {
		logger.Error("Failed to load i18n", "err", fmt.Errorf("failed to read translations: %w", err))
		os.Exit(1)
	}

	var data map[string]map[string]string
	if err := json.Unmarshal(file, &data); err != nil {
		logger.Error("Failed to load i18n", "err", fmt.Errorf("failed to unmarshal json: %w", err))
		os.Exit(1)
	}

	return &Translator{
		data:         data,
		fallbackLang: fallbackLang,
	}
}

/* Translates a key to the given language, with optional parameter replacement */
func (t *Translator) T(lang string, key string, param string) string {
	val := t.getRaw(lang, key)
	if val == "" {
		val = t.getRaw(t.fallbackLang, key)
		if val == "" {
			// Fallback to key if no translation found
			val = key
			// Strip err_ prefix and replace underscores with spaces
			if len(val) > 4 && val[:4] == "err_" {
				val = val[4:]
			}
			val = strings.ReplaceAll(val, "_", " ")
		}
	}

	if param == "" {
		return val
	}

	// Parameter replacement if not empty
	return strings.Replace(val, "{p}", param, 1)
}

/* Returns the raw translation string for the given language and key */
func (t *Translator) getRaw(lang, key string) string {
	if l, ok := t.data[lang]; ok {
		return l[key]
	}
	return ""
}

/* Insert user lang to context */
func InsertLanguageToContext(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, userLangKey, lang)
}

/* Get user lang from context */
func GetLanguageFromContext(ctx context.Context) string {
	lang, ok := ctx.Value(userLangKey).(string)
	if !ok {
		return "en"
	}
	return lang
}
