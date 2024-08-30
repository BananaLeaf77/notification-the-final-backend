package config

import (
	"fmt"
	"net/smtp"
	"os"
)

var smtpValue *smtp.Auth

func InitMailSMTP() (*smtp.Auth, error) {
	emailSender, err := getSender()
	if err != nil {
		return nil, err
	}

	emailPassword, err := getPassword()
	if err != nil {
		return nil, err
	}

	smtpHost, err := getHost()
	if err != nil {
		return nil, err
	}

	smtpV := smtp.PlainAuth("", *emailSender, *emailPassword, *smtpHost)

	if smtpV == nil {
		return nil, fmt.Errorf("failed to init SMTP")
	}

	smtpValue = &smtpV

	return smtpValue, nil
}

func getSender() (*string, error) {
	sender := os.Getenv("EMAIL_SENDER")
	if sender == "" {
		return nil, fmt.Errorf("email sender invalid, value : %s", sender)
	}
	return &sender, nil
}

func getHost() (*string, error) {
	sender := os.Getenv("SMTP_HOST")
	if sender == "" {
		return nil, fmt.Errorf("smtp value invalid, value : %s", sender)
	}
	return &sender, nil
}

func getPassword() (*string, error) {
	sender := os.Getenv("EMAIL_SENDER_PASSWORD")
	if sender == "" {
		return nil, fmt.Errorf("email sender password invalid, value : %s", sender)
	}
	return &sender, nil
}
