// repository/mailgun_repository.go

package repository

import (
	"context"
	"fmt"
	"notification/domain"
	"strconv"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type mailgunRepository struct {
	client      *mailgun.MailgunImpl
	senderEmail string
	schoolPhone string
}

func NewMailgunRepository(client *mailgun.MailgunImpl, schoolPhone string) domain.MailGunRepo {
	return &mailgunRepository{
		client:      client,
		schoolPhone: schoolPhone,
	}
}

func (m *mailgunRepository) SendMass(ctx context.Context, studName string, parentName string, emailAddress string) error {
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
	subject := fmt.Sprintf("Pemberitahuan Ketidakhadiran %s pada %s %s, tanggal %s", studName, hourAndMinute, isAM, formattedDate)

	// Email body
	body := fmt.Sprintf(`Kepada Yth. Bapak/Ibu %s,

Kami ingin memberitahukan bahwa putra/putri Bapak/Ibu, %s, tidak hadir di sekolah pada tanggal %s pukul %s %s.

Alasan ketidakhadiran belum kami terima hingga saat ini. Kami berharap Bapak/Ibu dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi putra/putri Bapak/Ibu.

Jika terdapat pertanyaan atau memerlukan bantuan lebih lanjut, Bapak/Ibu dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, parentName, studName, formattedDate, hourAndMinute, isAM, m.schoolPhone)

	// Create a new Mailgun message
	message := m.client.NewMessage(m.senderEmail, subject, body, emailAddress)

	// Set a timeout for the request context
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// Send the email
	_, _, err = m.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

