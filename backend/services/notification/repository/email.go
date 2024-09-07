package repository

import (
	"context"
	"fmt"
	"net/smtp"
	"notification/config"
	"notification/domain"
	"strconv"
	"time"
)

type emailSMTPRepository struct {
	client      smtp.Auth
	emailSender string
	schoolPhone string
	smtpAdress  string
}

func NewEmailSMTPRepository(client smtp.Auth, smtpAddress, schoolPhone, emailSender string) domain.EmailSMTPRepo {
	return &emailSMTPRepository{
		client:      client,
		emailSender: emailSender,
		schoolPhone: schoolPhone,
		smtpAdress:  smtpAddress,
	}
}

func (m *emailSMTPRepository) SendMass(ctx context.Context, payloadList *[]domain.EmailSMTPData) error {
	wg := config.GetWaitGroupInstance()

	errorChan := make(chan error, len(*payloadList))

	for _, payload := range *payloadList {
		wg.Add(1) 
		go func(payload domain.EmailSMTPData) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			// Prepare the email message
			if err := m.sendEmail(payload); err != nil {
				errorChan <- fmt.Errorf("failed to send email to %s: %w", payload.EmailAddress, err)
			}
		}(payload)
	}

	// Wait for all email-sending goroutines to complete
	wg.Wait()
	close(errorChan) // Close the error channel

	// Collect errors if any
	var err error
	for e := range errorChan {
		if e != nil {
			err = e
		}
	}
	return err
}

func (m *emailSMTPRepository) sendEmail(payload domain.EmailSMTPData) error {
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

	// Email subject
	subject := fmt.Sprintf("Pemberitahuan Ketidakhadiran %s pada %s %s, tanggal %s", payload.StudentName, hourAndMinute, isAM, formattedDate)

	// Email body
	body := fmt.Sprintf(`Kepada Yth. Bapak/Ibu %s,

Kami ingin memberitahukan bahwa putra/putri Bapak/Ibu, %s, tidak hadir di sekolah pada tanggal %s pukul %s %s.

Alasan ketidakhadiran belum kami terima hingga saat ini. Kami berharap Bapak/Ibu dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi putra/putri Bapak/Ibu.

Jika terdapat pertanyaan atau memerlukan bantuan lebih lanjut, Bapak/Ibu dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, payload.ParentName, payload.StudentName, formattedDate, hourAndMinute, isAM, m.schoolPhone)

	// Prepare the message
	msg := "From: " + m.emailSender + "\n" +
		"To: " + payload.EmailAddress + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	// Send the email
	err = smtp.SendMail(m.smtpAdress, m.client, m.emailSender, []string{payload.EmailAddress}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
