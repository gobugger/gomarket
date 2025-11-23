package localizer

import (
	"github.com/gobugger/gomarket/internal/translations"
	"testing"
)

func TestEscape(t *testing.T) {
	l, _ := Get(translations.En)

	if res := l.Translate(`We take 5%% cut`); res != "We take 5% cut" {
		t.Error(res)
	}
}
