package support

import (
	"github.com/gobugger/globalize"
	"github.com/gobugger/gomarket/internal/config"
	"time"
)

type Faq struct {
	Question string
	Answer   string
}

func GetFaqs(l *globalize.Localizer) []Faq {
	faqs := []Faq{
		{
			Question: l.Translate("How do I become a vendor?"),
			Answer:   l.Translate(`Create an account with two-factor authentication (2FA) enabled, then submit a vendor application.`),
		},
		{
			Question: l.Translate("How do I place an order?"),
			Answer: l.Translate(`Add the desired products to your cart. Navigate to /cart and select “Checkout,” choose your preferred delivery method, and provide the required information.
If you have available balance, you may use it at checkout. Otherwise, you may request an invoice. Invoices must be paid within %d hours for the order to continue processing.`, config.InvoicePaymentWindow/time.Hour),
		},
		{
			Question: l.Translate("How can I obtain Monero (XMR)?"),
			Answer:   l.Translate(`There are several ways to acquire Monero. Please visit getmonero.org for official guidance and resources.`),
		},
		{
			Question: l.Translate("What should I do if my order has not arrived or does not match the description?"),
			Answer: l.Translate(`Orders are automatically finalized after %d days.
You may extend the auto-finalization period twice.
If the order still has not arrived after these extensions, you should open a dispute.
Our team will review the case and make a resolution based on the available information.`, config.OrderDeliveryWindow/(time.Hour*24)),
		},
		{
			Question: l.Translate("What am I allowed to sell on this marketplace?"),
			Answer:   l.Translate(`Only legal items are permitted. If you are unsure whether your product qualifies, please contact us before creating a listing.`),
		},
	}
	return faqs
}
