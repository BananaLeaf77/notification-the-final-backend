package config

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/lib/pq"
	"github.com/skip2/go-qrcode"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

var (
	meowWhatsapp *whatsmeow.Client
	qrCodeSent   bool
	mu           sync.Mutex
)

func InitSender() (*whatsmeow.Client, smtp.Auth, *string, *string, *string, error) {
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

	fmt.Println("SMTP initialized")

	//Meow
	dbms, err := getDBMS()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	user, err := getDBUser()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	pass, err := getDBPassword()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	dbname, err := getDBName()
	if err != nil {
		return nil, nil, nil, nil, nil, err
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
	mClient := whatsmeow.NewClient(deviceStore, nil)
	meowWhatsapp = mClient

	if meowWhatsapp.Store.ID == nil {
		qrChan, _ := meowWhatsapp.GetQRChannel(context.Background())
		err = meowWhatsapp.Connect()
		if err != nil {
			panic(err)
		}

		// Process QR code
		for evt := range qrChan {
			if evt.Event == "code" {
				mu.Lock()
				if !qrCodeSent {
					fmt.Println("")
					fmt.Println("IMPORTANT no WhatsApp session was found !!")
					fmt.Println("Need admin to scan the QR code for the server to run properly!")
					// fmt.Println("==============   QR CODE   ==============")
					// fmt.Println(evt.Code)
					fmt.Println("Loading...")

					err := generateQRCode(evt.Code, "qrcode.png")
					if err != nil {
						panic(err)
					}

					err = SendQRtoEmail(smtpAddr, &smtpAuth, *emailSender, "qrcode.png")
					if err != nil {
						panic(err)
					}
					fmt.Printf("Image of QR Code is sent to %s, go ahead and scan them :)\n", *emailSender)
					fmt.Println("")

					qrCodeSent = true
				}
				mu.Unlock()
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err = meowWhatsapp.Connect()
		if err != nil {
			panic(err)
		}
		fmt.Println("WhatsMeow initialized")
	}

	return meowWhatsapp, smtpAuth, &smtpAddr, schoolPhone, emailSender, nil
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
		return nil, fmt.Errorf("DATABASE User is missing, value: %s", v)
	}
	return &v, nil
}

func getDBPassword() (*string, error) {
	v := os.Getenv("DB_PASSWORD")
	if v == "" {
		return nil, fmt.Errorf("DATABASE Password is missing, value: %s", v)
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

func generateQRCode(data, filePath string) error {
	// Generate QR code and save as an image file
	err := qrcode.WriteFile(data, qrcode.Medium, 256, filePath)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %v", err)
	}
	return nil
}

func SendQRtoEmail(smtpAddr string, smtpAuth *smtp.Auth, emailSender string, qrFilePath string) error {
	// Subject and body of the email
	subject := "Subject: SINOAN QR Code Login\n"
	body := "Please find the attached QR code for login.\n\n"

	// Open the QR code file
	fileData, err := os.ReadFile(qrFilePath)
	if err != nil {
		return fmt.Errorf("failed to read QR code file: %v", err)
	}

	// Get the file name and MIME boundary
	fileName := filepath.Base(qrFilePath)
	boundary := "my-boundary-12345"

	// Create the email header with MIME boundary
	msg := []byte("From: " + emailSender + "\n" +
		"To: " + emailSender + "\n" +
		subject +
		"MIME-Version: 1.0\n" +
		"Content-Type: multipart/mixed; boundary=" + boundary + "\n\n" +
		"--" + boundary + "\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\n\n" +
		body + "\n\n" +
		"--" + boundary + "\n" +
		"Content-Type: image/png\n" +
		"Content-Disposition: attachment; filename=\"" + fileName + "\"\n" +
		"Content-Transfer-Encoding: base64\n\n")

	// Encode the file content to base64 and append it to the message
	msg = append(msg, []byte(encodeBase64(fileData))...)
	msg = append(msg, []byte("\n--"+boundary+"--")...)

	// Send the email with the attachment
	err = smtp.SendMail(smtpAddr, *smtpAuth, emailSender, []string{emailSender}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// Helper function to encode file content to base64
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
