package config

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"os/exec"

	_ "github.com/lib/pq"

	"github.com/twilio/twilio-go"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

var meowWhatsapp *whatsmeow.Client

func InitSMTPEmailer() (smtp.Auth, *string, *string, *string, error) {
	// SMTP Emailer
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

	// Twillio
	twillioClient := twilio.NewRestClient()
	if twillioClient == nil {
		return nil, nil, nil, nil, fmt.Errorf("Failed to initialize twillio")
	}

	return smtpAuth, &smtpAddr, schoolPhone, emailSender, nil
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

func getDBMS() (*string, error) {
	dbms := os.Getenv("DBMS")
	if dbms == "" {
		return nil, fmt.Errorf("DBMS is missing, value: %s", dbms)
	}
	return &dbms, nil
}

func getDBUser() (*string, error) {
	v := os.Getenv("DB_USER")
	if v == "" {
		return nil, fmt.Errorf("Database User is missing, value: %s", v)
	}
	return &v, nil
}

func getDBPassword() (*string, error) {
	v := os.Getenv("DB_PASSWORD")
	if v == "" {
		return nil, fmt.Errorf("Database Password is missing, value: %s", v)
	}
	return &v, nil
}

func getDBName() (*string, error) {
	v := os.Getenv("DB_DATABASE")
	if v == "" {
		return nil, fmt.Errorf("DB Name is missing, value: %s", v)
	}
	return &v, nil
}

func InitMeow() (*whatsmeow.Client, error) {

	dbms, err := getDBMS()
	if err != nil {
		return nil, err
	}

	user, err := getDBUser()
	if err != nil {
		return nil, err
	}

	pass, err := getDBPassword()
	if err != nil {
		return nil, err
	}

	dbname, err := getDBName()
	if err != nil {
		return nil, err
	}

	meowAddress := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", *user, *pass, *dbname)

	container, err := sqlstore.New(*dbms, meowAddress, nil)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	client := whatsmeow.NewClient(deviceStore, nil)
	meowWhatsapp = client

	if meowWhatsapp.Store.ID == nil {
		qrChan, _ := meowWhatsapp.GetQRChannel(context.Background())
		err = meowWhatsapp.Connect()
		if err != nil {
			panic(err)
		}
		// If there is no stored session, show the QR code to log in.
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("")
				fmt.Println("Need admin to scan the qr code for the server to run properly!")
				fmt.Println("==============   QR CODE   ==============")
				fmt.Println(evt.Code)
				fmt.Println("QR CODE image is sent to @dognub61@gmail.com, go ahead and scan them :)")

				err := generateQRCode(evt.Code, "qrcode.png")
				if err != nil {
					panic(err)
				}

				fmt.Println("")
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err = meowWhatsapp.Connect()
		if err != nil {
			panic(err)
		}
		fmt.Println("Login success")
	}

	return meowWhatsapp, nil
}

func generateQRCode(data, filePath string) error {
	// Run the qrencode command to generate the QR code as an image
	cmd := exec.Command("qrencode", "-o", filePath, data)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %v", err)
	}
	return nil
}
