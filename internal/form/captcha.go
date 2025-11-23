package form

import (
	"github.com/gobugger/gomarket/internal/captcha"
)

type CaptchaValidable interface {
	Answer() captcha.Answer
}

type Captcha struct {
	CaptchaAnswer captcha.Answer `schema:"captcha_answer" validate:"required"`
}

func (c *Captcha) Answer() captcha.Answer {
	return c.CaptchaAnswer
}
