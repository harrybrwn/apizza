package dawg

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Card is an interface representing a credit or debit card.
type Card interface {
	// Number should return the card number.
	Num() string

	// ExpiresOn returns the date that the payment exprires.
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

var badExpiration = time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)

// ExpiresOn returns the expiration date as a time.Time.
func (p *Payment) ExpiresOn() time.Time {
	m, y := parseDate(p.Expiration)
	if m < 0 || y < 0 {
		return badExpiration
	}
	return time.Date(int(y), time.Month(m), 1, 0, 0, 0, 0, time.Local)
}

// Code returns the CVV.
func (p *Payment) Code() string {
	return p.CVV
}

var _ Card = (*Payment)(nil)

func makeOrderPaymentFromCard(c Card) *orderPayment {
	return &orderPayment{
		Number:       c.Num(),
		Expiration:   formatDate(c.ExpiresOn()),
		SecurityCode: c.Code(),
		Type:         "CreditCard",
	}
}

func formatDate(t time.Time) string {
	year := fmt.Sprintf("%d", t.Year())
	if len(year) >= 4 {
		year = year[len(year)-2:]
	}
	return fmt.Sprintf("%02d%s", t.Month(), year)
}

func parseDate(d string) (month int, year int) {
	parts := strings.Split(d, "/")
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
	ProviderID     string
	OTP            string
	GpmPaymentType string `json:"gpmPaymentType"`
}
