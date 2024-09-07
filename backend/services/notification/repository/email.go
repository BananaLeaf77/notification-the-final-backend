package repository

import (
	"context"
	"fmt"
	"net/smtp"
	"notification/domain"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type emailSMTPRepository struct {
	db          *pgxpool.Pool
	client      smtp.Auth
	emailSender string
	schoolPhone string
	smtpAdress  string
}

func NewEmailSMTPRepository(db *pgxpool.Pool, client smtp.Auth, smtpAddress, schoolPhone, emailSender string) domain.EmailSMTPRepo {
	return &emailSMTPRepository{
		db:          db,
		client:      client,
		emailSender: emailSender,
		schoolPhone: schoolPhone,
		smtpAdress:  smtpAddress,
	}
}

func (m *emailSMTPRepository) SendMass(ctx context.Context, idList *[]int) error {
	var finalErr error

	for _, id := range *idList {
		student, err := m.fetchStudentDetails(ctx, id)
		if err != nil {
			finalErr = fmt.Errorf("failed to fetch student details for ID %d: %w", id, err)
			continue
		}

		if err := m.sendEmail(student); err != nil {
			finalErr = fmt.Errorf("failed to send email to %s: %w", student.Parent.Email, err)
			continue
		}
	}

	return finalErr
}

// Fetches student details including parent email from the database.
func (m *emailSMTPRepository) fetchStudentDetails(ctx context.Context, studentID int) (domain.EmailSMTPData, error) {
	var stuctHolder domain.EmailSMTPData

	// SQL query to fetch student and parent details.
	query := `
		SELECT s.name, p.email, p.name, p.gender
		FROM students s
		JOIN parents p ON s.parent_id = p.id 
		WHERE s.id = $1 AND s.deleted_at IS NULL AND p.deleted_at IS NULL;
	`

	// Execute the query and scan the result into the student structure.
	err := m.db.QueryRow(ctx, query, studentID).Scan(
		&stuctHolder.Student.Name,
		&stuctHolder.Parent.Email,
		&stuctHolder.Parent.Name,
		&stuctHolder.Parent.Gender,
	)

	if err != nil {
		return domain.EmailSMTPData{}, fmt.Errorf("could not fetch student details: %v", err)
	}

	return stuctHolder, nil
}

// Sends an email to the provided email address.
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

	// Email subject.
	subject := fmt.Sprintf("Pemberitahuan Ketidakhadiran %s pada %s %s, tanggal %s", payload.Student.Name, hourAndMinute, isAM, formattedDate)

	// Email body.
	bodyMale := fmt.Sprintf(`Kepada Yth. Bapak %s,

Kami ingin memberitahukan bahwa anak Bapak, %s, tidak hadir di sekolah pada tanggal %s pukul %s %s.

Alasan ketidakhadiran belum kami terima hingga saat ini. Kami berharap Bapak dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi anak Bapak.

Jika terdapat pertanyaan atau memerlukan bantuan lebih lanjut, Bapak dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, payload.Parent.Name, payload.Student.Name, formattedDate, hourAndMinute, isAM, m.schoolPhone)

	bodyFemale := fmt.Sprintf(`Kepada Yth. Ibu %s,

Kami ingin memberitahukan bahwa anak Ibu, %s, tidak hadir di sekolah pada tanggal %s pukul %s %s.

Alasan ketidakhadiran belum kami terima hingga saat ini. Kami berharap Ibu dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi anak Ibu.

Jika terdapat pertanyaan atau memerlukan bantuan lebih lanjut, Ibu dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, payload.Parent.Name, payload.Student.Name, formattedDate, hourAndMinute, isAM, m.schoolPhone)

	var msg string

	if payload.Parent.Gender == "Female" {
		msg = "From: " + m.emailSender + "\n" +
			"To: " + payload.Parent.Email + "\n" +
			"Subject: " + subject + "\n\n" +
			bodyFemale
	} else if payload.Parent.Gender == "Male" {
		msg = "From: " + m.emailSender + "\n" +
			"To: " + payload.Parent.Email + "\n" +
			"Subject: " + subject + "\n\n" +
			bodyMale
	}

	// Send the email.
	err = smtp.SendMail(m.smtpAdress, m.client, m.emailSender, []string{payload.Parent.Email}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
