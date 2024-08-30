package config

import (
	"fmt"
	"net/smtp"
	"os"
)

func InitMailSMTP() (smtp.Auth, *string, *string, *string, error) {
	emailSender, err := getSender()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	emailPassword, err := getPassword()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	smtpHost, err := getHost()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	smtpPort, err := getSMTPPort()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	schoolPhone, err := getSchoolPhone()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	smtpAuth := smtp.PlainAuth("", *emailSender, *emailPassword, *smtpHost)

	smtpAddr := fmt.Sprintf("%s:%s", *smtpHost, *smtpPort)

	return smtpAuth, &smtpAddr, schoolPhone, emailSender, nil
}

func getSender() (*string, error) {
	sender := os.Getenv("EMAIL_SENDER")
	if sender == "" {
		return nil, fmt.Errorf("email sender invalid, value : %s", sender)
	}
	return &sender, nil
}

func getHost() (*string, error) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return nil, fmt.Errorf("smtp value invalid, value : %s", host)
	}
	return &host, nil
}

func getPassword() (*string, error) {
	pass := os.Getenv("EMAIL_SENDER_PASSWORD")
	if pass == "" {
		return nil, fmt.Errorf("email password invalid, value : %s", pass)
	}
	return &pass, nil
}

func getSchoolPhone() (*string, error) {
	phone := os.Getenv("SCHOOL_PHONE")
	if phone == "" {
		return nil, fmt.Errorf("school phone invalid, value : %s", phone)
	}
	return &phone, nil
}

func getSMTPPort() (*string, error) {
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		return nil, fmt.Errorf("smtp port invalid, value : %s", port)
	}
	return &port, nil
}
