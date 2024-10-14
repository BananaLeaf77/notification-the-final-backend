package repository

import (
	"context"
	"fmt"
	"net/smtp"
	"notification/domain"
	"strconv"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

// init var
var (
	bodyMale   string
	bodyFemale string
	tNow       time.Time
	subject    string
	JID        types.JID
)

type senderRepository struct {
	db          *gorm.DB
	client      smtp.Auth
	emailSender string
	schoolPhone string
	smtpAdress  string
	meowClient  *whatsmeow.Client
}

func NewSenderRepository(db *gorm.DB, client smtp.Auth, smtpAddress, schoolPhone, emailSender string, meow *whatsmeow.Client) domain.SenderRepo {
	return &senderRepository{
		db:          db,
		client:      client,
		emailSender: emailSender,
		schoolPhone: schoolPhone,
		smtpAdress:  smtpAddress,
		meowClient:  meow,
	}
}

func (m *senderRepository) SendMass(ctx context.Context, idList *[]int) error {
	var finalErr error
	tNow = time.Now()

	for _, id := range *idList {
		student, err := m.fetchStudentDetails(ctx, id)
		if err != nil {
			finalErr = fmt.Errorf("failed to fetch student details for ID %d: %w", id, err)
			continue
		}

		err = m.initText(student)
		if err != nil {
			return err
		}

		if student.Parent.Email != nil && *student.Parent.Email != "" {
			if err := m.sendEmail(student); err != nil {
				finalErr = fmt.Errorf("failed to send email to %s: %w", *student.Parent.Email, err)
				continue
			}

			// Add to history table
		}

		err = m.sendWA(ctx, student)
		if err != nil {
			finalErr = fmt.Errorf("failed to send Whatsapp text to %s: %w", student.Parent.Telephone, err)
			continue
		}
	}

	return finalErr
}

func (m *senderRepository) fetchStudentDetails(ctx context.Context, studentID int) (*domain.StudentAndParent, error) {
	var student domain.Student
	var parent domain.Parent

	err := m.db.WithContext(ctx).Where("student_id = ?", studentID).Preload("Parent").First(&student).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("student with ID %d not found", studentID)
		}
		return nil, fmt.Errorf("could not fetch student details: %v", err)
	}

	err = m.db.WithContext(ctx).Where("parent_id = ?", student.ParentID).First(&parent).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("parent with ID %d not found", student.ParentID)
		}
		return nil, fmt.Errorf("could not fetch parent details: %v", err)
	}

	return &domain.StudentAndParent{
		Student: student,
		Parent:  parent,
	}, nil
}

// Sends an email to the provided email address.
func (m *senderRepository) sendEmail(payload *domain.StudentAndParent) error {
	var msg string

	if payload.Parent.Gender == "female" {
		msg = "From: " + m.emailSender + "\n" +
			"To: " + *payload.Parent.Email + "\n" +
			"Subject: " + subject + "\n\n" +
			bodyFemale
	} else if payload.Parent.Gender == "male" {
		msg = "From: " + m.emailSender + "\n" +
			"To: " + *payload.Parent.Email + "\n" +
			"Subject: " + subject + "\n\n" +
			bodyMale
	}

	// Send the email.
	err := smtp.SendMail(m.smtpAdress, m.client, m.emailSender, []string{*payload.Parent.Email}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (m *senderRepository) sendWA(ctx context.Context, payload *domain.StudentAndParent) error {
	var msg string
	completeFormat := fmt.Sprintf("%s%s", "62", payload.Parent.Telephone[1:])

	jid := types.NewJID(completeFormat, types.DefaultUserServer)

	if payload.Parent.Gender == "female" {
		msg = bodyFemale
	} else if payload.Parent.Gender == "male" {
		msg = bodyMale
	}

	conversationMessage := &waE2E.Message{
		Conversation: &msg,
	}

	_, err := m.meowClient.SendMessage(ctx, jid, conversationMessage)
	if err != nil {
		fmt.Println("error cuk")
		return err
	}
	return nil
}

func (m *senderRepository) initText(payload *domain.StudentAndParent) error {
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

	subject = fmt.Sprintf("Pemberitahuan Ketidakhadiran %s pada %s %s, tanggal %s", payload.Student.Name, hourAndMinute, isAM, formattedDate)

	bodyMale = fmt.Sprintf(`Kepada Yth. Bapak %s,

Kami ingin memberitahukan bahwa anak Bapak, %s, tidak hadir di sekolah pada tanggal %s pukul %s %s.

Alasan ketidakhadiran belum kami terima hingga saat ini. Kami berharap Bapak dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi anak Bapak.

Jika terdapat pertanyaan atau memerlukan bantuan lebih lanjut, Bapak dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, payload.Parent.Name, payload.Student.Name, formattedDate, hourAndMinute, isAM, m.schoolPhone)

	bodyFemale = fmt.Sprintf(`Kepada Yth. Ibu %s,

Kami ingin memberitahukan bahwa anak Ibu, %s, tidak hadir di sekolah pada tanggal %s pukul %s %s.

Alasan ketidakhadiran belum kami terima hingga saat ini. Kami berharap Ibu dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi anak Ibu.

Jika terdapat pertanyaan atau memerlukan bantuan lebih lanjut, Ibu dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, payload.Parent.Name, payload.Student.Name, formattedDate, hourAndMinute, isAM, m.schoolPhone)

	return nil
}
