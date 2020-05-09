package dawg

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Card is an interface representing a credit or debit card.
type Card interface {
	// Number should return the card number.
	Num() string

	// ExpiresOn returns the date that the payment expires.
	ExpiresOn() time.Time

	// Code returns the security code or the cvv.
	Code() string
}

// NewCard will create a new Card objected. If the expiration format is wrong then
// it will return nil. The expiration format should be "mm/yy".
func NewCard(number, expiration string, cvv int) Card {
	if len(expiration) < 4 || len(expiration) > 5 {
		return nil // bad expiration date format
	}

	return &Payment{
		Number:     number,
		Expiration: expiration,
		CVV:        strconv.Itoa(cvv),
	}
}

// ToPayment converts a card interface to a Payment struct.
func ToPayment(c Card) *Payment {
	return &Payment{
		Number:     c.Num(),
		Expiration: formatDate(c.ExpiresOn()),
		CVV:        c.Code(),
	}
}

// Payment just a way to compartmentalize a payment sent to dominos.
type Payment struct {
	// Number is the card number.
	Number string `json:"Number"`

	// Expiration is the expiration date of the card formatted exactly as
	// it is on the physical card.
	Expiration string `json:"Expiration"`
	CardType   string `json:"Type"`
	CVV        string `json:"SecurityCode"`
}

// Num returns the card number as a string.
func (p *Payment) Num() string {
	return p.Number
}

// BadExpiration is a zero datetime object the is returned on error.
var BadExpiration = time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)

// ExpiresOn returns the expiration date as a time.Time.
func (p *Payment) ExpiresOn() time.Time {
	if len(p.Expiration) == 0 {
		return BadExpiration
	}

	m, y := parseDate(p.Expiration)
	if m <= 0 || m > 12 || y < 0 {
		return BadExpiration
	}
	return time.Date(int(y), time.Month(m), 1, 0, 0, 0, 0, time.Local)
}

// Code returns the CVV.
func (p *Payment) Code() string {
	return p.CVV
}

// ValidateCard will return an error if the card given has any bad data.
func ValidateCard(c Card) error {
	if BadExpiration.Equal(c.ExpiresOn()) {
		return errors.New("card has a bad expiration date format")
	}
	if findCardType(c.Num()) == "" {
		return errors.New("could not find card ")
	}
	return nil
}

var _ Card = (*Payment)(nil)

func makeOrderPaymentFromCard(c Card) *orderPayment {
	return &orderPayment{
		Number:       c.Num(),
		Expiration:   formatDate(c.ExpiresOn()),
		SecurityCode: c.Code(),
		Type:         "CreditCard",
		CardType:     findCardType(c.Num()),
	}
}

func formatDate(t time.Time) string {
	year := fmt.Sprintf("%d", t.Year())
	if len(year) >= 4 {
		year = year[len(year)-2:]
	}
	return fmt.Sprintf("%02d%s", t.Month(), year)
}

// in the future, i may use `time.Parse("2/06", dateString)`
// and then try that with a few different date formats like "2/2006" or "02-06"
func parseDate(d string) (month int, year int) {
	parts := strings.Split(d, "/")

	if len(parts) == 1 && len(d) == 4 {
		// we have been given mmYY instead of mm/YY
		parts = []string{d[:2], d[2:]}
	}
	if len(parts) != 2 {
		return -1, -1
	}

	m, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return -1, -1
	}
	if len(parts[1]) < 4 {
		parts[1] = "20" + parts[1] // the first two digits will be truncated anyways
	}
	y, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return -1, -1
	}
	return int(m), int(y)
}

var cardRegex = map[string]*regexp.Regexp{
	"AmericanExpress": regexp.MustCompile(`^(?:3[47][0-9]{13})$`),
	"Discover":        regexp.MustCompile(`^6(?:011|5[0-9]{2})[0-9]{12}$`),
	"JCB":             regexp.MustCompile(`^(?:(?:2131|1800|35\d{3})\d{11})$`),
	"Maestro":         regexp.MustCompile(`^(?:(?:5[0678]\d\d|6304|6390|67\d\d)\d{8,15})$`),
	"MasterCard":      regexp.MustCompile(`^(?:(?:5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12})$`),
	"DinersClub":      regexp.MustCompile(`^3(?:0[0-5]|[68][0-9])[0-9]{11}$`),
	"Visa":            regexp.MustCompile(`^4[0-9]{12}(?:[0-9]{3})?$`),
	"Enroute":         regexp.MustCompile(`^(?:2014|2149)\d{11}$`),
}

func findCardType(num string) string {
	for ctype, pat := range cardRegex {
		if pat.MatchString(num) {
			return ctype
		}
	}
	return ""
}

// this is the struct that will actually be turning into json an will
// be sent to dominos.
type orderPayment struct {
	Number       string
	Expiration   string
	SecurityCode string
	Type         string
	CardType     string
	PostalCode   string

	// These next fields are just for dominos

	Amount         float64
	CardID         string `json:"CardID,omitempty"`
	ProviderID     string
	OTP            string
	GpmPaymentType string `json:"gpmPaymentType,omitempty"`
}
