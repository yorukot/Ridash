package email

import (
	"fmt"
	"net/mail"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateAddress checks the address with go-playground/validator and returns the normalized address.
func ValidateAddress(address string) (string, error) {
	if err := validate.Var(address, "required,email"); err != nil {
		return "", fmt.Errorf("invalid email address %q: %w", address, err)
	}

	parsed, err := mail.ParseAddress(address)
	if err != nil {
		return "", fmt.Errorf("invalid email address %q: %w", address, err)
	}
	return parsed.Address, nil
}
