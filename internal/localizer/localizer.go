package localizer

import (
	"github.com/gobugger/gomarket/internal/translations"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Localizer struct {
	ID      string
	printer *message.Printer
}

var localizers = []Localizer{}

func init() {
	for _, locale := range translations.Locales() {
		localizers = append(localizers, Localizer{
			ID:      locale,
			printer: message.NewPrinter(language.MustParse(locale)),
		})
	}
}

// TODO: Fallback to closest language if available
func Get(lang string) (Localizer, bool) {
	for _, l := range localizers {
		if lang == l.ID {
			return l, true
		}
	}

	return localizers[0], false
}

// We also add a Translate() method to the Localizer type. This acts
// as a wrapper around the unexported message.Printer's Sprintf()
// function and returns the appropriate translation for the given
// message and arguments.
func (l Localizer) Translate(key message.Reference, args ...any) string {
	return l.printer.Sprintf(key, args...)
}
