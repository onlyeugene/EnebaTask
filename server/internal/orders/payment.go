package orders

import (
	"strconv"
	"strings"
	"time"
)

// DeclineCard is the one card number the simulated payment turns down (API.md).
const DeclineCard = "4000000000000002"

// checkCard checks the card's format only; there is no real payment provider. It returns the
// number without spaces or dashes. No Luhn check: "any valid-looking number pays" (API.md),
// and a tester typing 16 random digits shouldn't be turned away.
func checkCard(c Card, now time.Time) (string, error) {
	fields := map[string]string{}

	number := strings.NewReplacer(" ", "", "-", "").Replace(c.CardNumber)
	if len(number) < 12 || len(number) > 19 || !allDigits(number) {
		fields["cardNumber"] = "must be 12 to 19 digits"
	}

	month, year, ok := parseExpiry(c.Expiry)
	switch {
	case !ok:
		fields["expiry"] = "must be MM/YY"
	// A card works until the end of its expiry month.
	case !now.Before(time.Date(2000+year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC)):
		fields["expiry"] = "card has expired"
	}

	if cvc := strings.TrimSpace(c.CVC); len(cvc) < 3 || len(cvc) > 4 || !allDigits(cvc) {
		fields["cvc"] = "must be 3 or 4 digits"
	}
	if name := strings.TrimSpace(c.Name); name == "" || len(name) > 100 {
		fields["name"] = "is required, at most 100 characters"
	}

	if len(fields) > 0 {
		return "", &ValidationError{Fields: fields}
	}
	return number, nil
}

// simulate is the fake payment provider: every card pays except DeclineCard.
func simulate(number string) string {
	if number == DeclineCard {
		return OutcomeDeclined
	}
	return OutcomeSuccess
}

// parseExpiry reads "MM/YY" (spaces around the slash allowed).
func parseExpiry(s string) (month, year int, ok bool) {
	mm, yy, found := strings.Cut(s, "/")
	mm, yy = strings.TrimSpace(mm), strings.TrimSpace(yy)
	if !found || len(mm) != 2 || len(yy) != 2 || !allDigits(mm) || !allDigits(yy) {
		return 0, 0, false
	}
	month, _ = strconv.Atoi(mm)
	year, _ = strconv.Atoi(yy)
	return month, year, month >= 1 && month <= 12
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
