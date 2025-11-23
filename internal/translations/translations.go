package translations

import (
	"golang.org/x/text/message"
	"slices"
	"sync"
)

// To support new language you need to do two things
// 1. Add new language to the -lang list below
// 2. Run: make translate && export OPENAI_API_KEY=... && go run ./tool/translator && make translate

var locales []string
var mtx sync.Mutex

const (
	DefaultLocale string = "en-US"
)

func Locales() []string {
	if locales == nil {
		mtx.Lock()
		defer mtx.Unlock()

		tags := message.DefaultCatalog.Languages()
		locales = make([]string, len(tags))

		for i, tag := range tags {
			locales[i] = tag.String()
		}
	}

	return slices.Clone(locales)
}

func ValidLocale(locale string) bool {
	return slices.Contains(locales, locale)
}

//go:generate go tool gotext -srclang=en-US update -out=catalog.go -lang=en-US,sv-SE,de-DE,fr-FR,lt-LT,fi-FI,et-EE,da-DK,no-NO,lv-LV,pl-PL,es-ES,pt-PT,it-IT github.com/gobugger/gomarket/cmd/market
