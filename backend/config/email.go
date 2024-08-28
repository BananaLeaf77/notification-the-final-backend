package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/gomail.v2"
)

var goMailDialer *gomail.Dialer

func InitEmailer() error {
	goMailDialer = gomail.NewDialer("mail.smtp2go.com", 2525, "your-username", "your-password")

	_, err := getSender()
	if err != nil {
		return err
	}

	_, err = getSchoolPhoneNumber()
	if err != nil {
		return err
	}

	return nil

}

func SendEmail(studName string, parentName, emailAddress string) error {
	goMailMessage := gomail.NewMessage()
	schoolPhone, _ := getSchoolPhoneNumber()
	senderEmail, err := getSender()
	if err != nil {
		return err
	}
	goMailMessage.SetHeader("From", senderEmail)

	goMailMessage.SetHeader("To", emailAddress)

	tNow := time.Now()
	formattedDate := tNow.Format("02/01/2006")
	hourOnly := tNow.Format("15")
	hourAndMinute := tNow.Format("15:04")

	intHourOnly, err := strconv.Atoi(hourOnly)
	if err != nil {
		return err
	}
	isAM := "AM"
	if intHourOnly >= 12 {
		isAM = "PM"
	}

	// email subject
	subject := fmt.Sprintf("Pemberitahuan Ketidakhadiran %s pada %s %s, tanggal %s", studName, hourAndMinute, isAM, formattedDate)
	goMailMessage.SetHeader("Subject", subject)

	// email body
	body := fmt.Sprintf(`Kepada Yth. Bapak/Ibu %s,

Kami ingin memberitahukan bahwa putra/putri Bapak/Ibu, %s, tidak hadir di sekolah pada tanggal %s pukul %s %s.

Alasan ketidakhadiran belum kami terima hingga saat ini. Kami berharap Bapak/Ibu dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi putra/putri Bapak/Ibu.

Jika terdapat pertanyaan atau memerlukan bantuan lebih lanjut, Bapak/Ibu dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, parentName, studName, formattedDate, hourAndMinute, isAM, schoolPhone)

	goMailMessage.SetBody("text/plain", body)

	// Send the email
	if err := goMailDialer.DialAndSend(goMailMessage); err != nil {
		return err
	}

	return nil
}

func getSender() (string, error) {
	emailSender := os.Getenv("EMAIL_SENDER")
	if emailSender == "" {
		return "", fmt.Errorf("empty email sender")
	}
	return emailSender, nil
}

func getSchoolPhoneNumber() (string, error) {
	sp := os.Getenv("SCHOOL_PHONE")
	if sp == "" {
		return "", fmt.Errorf("empty school phone number")
	}
	return sp, nil

}
