package main

import (
	"gopkg.in/gomail.v2"
)

func sendEmail() {
	m := gomail.NewMessage()
	m.SetHeader("From", "madegedearysutha@email.com")
	m.SetHeader("To", "dognub61@gmail.com")
	m.SetHeader("Subject", "Hello brodi!")
	m.SetBody("text/plain", "Pada suatu hari saya pun me ngetest library gomail.")

	d := gomail.NewDialer("mail.smtp2go.com", 2525, "your-username", "your-password")

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func main() {
	sendEmail()
}
