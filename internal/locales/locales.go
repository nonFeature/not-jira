package locales

import "strings"

// ForUser returns the appropriate localization bundle based on Telegram user's language code.
// If the user's language starts with "ru", Russian is returned.
// For all other languages, English is returned.
func ForUser(langCode string) *Bundle {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(langCode)), "ru") {
		return GetRu()
	}
	return GetEn()
}
