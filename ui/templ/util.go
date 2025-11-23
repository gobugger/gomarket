package ui

import (
	"fmt"
	"github.com/gobugger/gomarket/internal/localizer"
	"github.com/gobugger/gomarket/internal/qrcode"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"log/slog"
	"strings"
	"time"
)

func (tc *TemplateContext) Translate(data any, args ...any) string {
	message := ""
	switch data := data.(type) {
	case string:
		message = data
	case repo.OrderStatus:
		message = string(data)
	case repo.InvoiceStatus:
		message = string(data)
	case repo.WithdrawalStatus:
		message = string(data)
	default:
		slog.Error("Unknown data type", slog.Any("type", data))
		return "TRANSLATION ERROR"
	}

	l, _ := localizer.Get(tc.Settings.Lang)
	return l.Translate(message, args...)
}

func (tc *TemplateContext) XMR2Fiat(pico int64) string {
	return fmt.Sprintf("%.2f", float64(currency.XMR2Fiat(currency.Currency(tc.Settings.Currency), pico))/100.0)
}

func (tc *TemplateContext) XMR2Fiat2(pico int64) string {
	c := currency.Currency(tc.Settings.Currency)
	return fmt.Sprintf("%s%.2f", c.Symbol(), float64(currency.XMR2Fiat(c, pico))/100.0)
}

func (tc *TemplateContext) Fiat(cents int64) string {
	return fmt.Sprintf("%.2f", float64(currency.Fiat2Fiat(currency.DefaultCurrency, currency.Currency(tc.Settings.Currency), cents))/100)
}

func (tc *TemplateContext) Fiat2(cents int64) string {
	c := currency.Currency(tc.Settings.Currency)
	return fmt.Sprintf("%s%.2f", c.Symbol(), float64(currency.Fiat2Fiat(currency.DefaultCurrency, c, cents))/100)
}

func (tc *TemplateContext) PricePerUnit(priceCent int64, units int32) string {
	c := currency.Currency(tc.Settings.Currency)
	priceCent = currency.Fiat2Fiat(currency.Currency(tc.Settings.Currency), currency.DefaultCurrency, priceCent)
	return fmt.Sprintf("%s%.2f", c.Symbol(), float64(priceCent)/float64(units)/100)
}

func FmtTime(t time.Time) string {
	return t.Format("15:04:05 02-01-2006")
}

func FmtDate(t time.Time) string {
	return t.Format("02-01-2006")
}

func FmtDuration(t time.Duration) string {
	if t < time.Duration(0) {
		return "00:00:00:00"
	}
	days := t / (time.Hour * 24)
	hours := (t - days*time.Hour*24) / time.Hour
	minutes := (t - days*time.Hour*24 - hours*time.Hour) / time.Minute
	seconds := (t - days*time.Hour*24 - hours*time.Hour - minutes*time.Minute) / time.Second

	res := ""
	if days > 0 {
		res += fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		res += fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		res += fmt.Sprintf("%dm", minutes)
	}
	res += fmt.Sprintf("%ds", seconds)

	return res
}

func (td *TemplateContext) FmtSince(t time.Time) string {
	l, _ := localizer.Get(td.Settings.Lang)

	since := time.Since(t)
	if since < 0 {
		return l.Translate("0 seconds")
	} else if since < time.Minute {
		return l.Translate("%d seconds", since/time.Second)
	} else if since < time.Hour {
		return l.Translate("%d minutes", since/time.Minute)
	} else if since < 24*time.Hour {
		return l.Translate("%d hours", since/time.Hour)
	} else {
		return l.Translate("%d days", since/(24*time.Hour))
	}
}

func DefaultValue(form Form, fieldName string) string {
	return form.Values[fieldName]
}

// Retuns the head of text up to maxlen, cut from the last space
func Head(text string, maxlen int) string {
	runes := []rune(text)
	if len(runes) <= maxlen {
		return text
	}

	result := string(runes[:maxlen])
	return result + "..."
}

func IsOnline(prevLogin time.Time) bool {
	return time.Since(prevLogin) < 12*time.Hour
}

func QrCodeSrc(content string) string {
	buf := &strings.Builder{}

	if _, err := buf.WriteString(`data:image/png;base64,`); err != nil {
		slog.Error("render qrcode", slog.Any("error", err))
		return ""
	}

	if err := qrcode.Base64Encode(buf, content); err != nil {
		slog.Error("render qrcode", slog.Any("error", err))
		return ""
	}

	return buf.String()
}
