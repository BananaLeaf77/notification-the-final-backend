package config

import (
	"fmt"
	"os"

	"github.com/mailgun/mailgun-go/v4"
)

func InitMailGun() (*mailgun.MailgunImpl, error) {
	mailGunDomain := os.Getenv("MAILGUN_DOMAIN")
	mailGunApiKey := os.Getenv("MAILGUN_API_KEY")

	mg := mailgun.NewMailgun(mailGunDomain, mailGunApiKey)
	if mg != nil{
		return nil, fmt.Errorf("failed to initialized Mailgun")
	}
	return mg, nil
}
