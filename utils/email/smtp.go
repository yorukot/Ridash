package email

import (
	"net"
	"net/smtp"
	"strconv"
)

// SMTPSettings holds connection and sender details for SMTP.
type SMTPSettings struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// NewSMTPSettings validates the sender address and constructs SMTP settings.
func NewSMTPSettings(host string, port int, username, password, from string) (*SMTPSettings, error) {
	validatedFrom, err := ValidateAddress(from)
	if err != nil {
		return nil, err
	}

	return &SMTPSettings{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     validatedFrom,
	}, nil
}

// Address returns the host:port string for the SMTP server.
func (s *SMTPSettings) Address() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
}

// Auth returns SMTP plain authentication using the configured credentials.
func (s *SMTPSettings) Auth() smtp.Auth {
	return smtp.PlainAuth("", s.Username, s.Password, s.Host)
}
