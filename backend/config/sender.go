package config

import (
	"fmt"
	"net/smtp"
	"os"

	"github.com/twilio/twilio-go"
)

func InitMessenger() (*twilio.RestClient, smtp.Auth, *string, *string, *string, error) {
	// SMTP Emailer
	emailSender, err := getSender()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	emailPassword, err := getPassword()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	smtpHost, err := getHost()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	smtpPort, err := getSMTPPort()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	schoolPhone, err := getSchoolPhone()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	smtpAuth := smtp.PlainAuth("", *emailSender, *emailPassword, *smtpHost)

	smtpAddr := fmt.Sprintf("%s:%s", *smtpHost, *smtpPort)

	// Twillio
	twillioClient := twilio.NewRestClient()
	if twillioClient == nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("Failed to initialize twillio")
	}

	return twillioClient, smtpAuth, &smtpAddr, schoolPhone, emailSender, nil
}

// For SMTP Emailer

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

// For Twillio

func getAccountSID() (*string, error) {
	sid := os.Getenv("TWILIO_ACCOUNT_SID")
	if sid == "" {
		return nil, fmt.Errorf("Twilio Account SID is missing, value: %s", sid)
	}
	return &sid, nil
}

func getAuthToken() (*string, error) {
	token := os.Getenv("TWILIO_AUTH_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("Twilio Auth Token is missing, value: %s", token)
	}
	return &token, nil
}

func getFromNumber() (*string, error) {
	number := os.Getenv("TWILIO_FROM_NUMBER")
	if number == "" {
		return nil, fmt.Errorf("Twilio From Number is missing, value: %s", number)
	}
	return &number, nil
}
