package repository

import (
	"context"
	"fmt"
	"net/smtp"
	"notification/domain"
	"strconv"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

// init var
var (
	bodyMale        string
	bodyFemale      string
	subject         string
	JID             types.JID
	bodyMaleEmail   string
	bodyFemaleEmail string

	bodyTestScore string
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

func (m *senderRepository) createTestScoreEmail(individual domain.IndividualExamScore, examType string) {
	// Start with a general introduction
	if individual.Student.Parent.Gender != "male" {
		body := fmt.Sprintf(`
		SINOAN Service 🔔
		
		Kepada Yth. Ibu %s,
		
		Kami ingin memberikan informasi mengenai hasil ujian %s anak Anda, %s Kelas %s. Berikut adalah detail hasil ujian pada beberapa mata pelajaran:
		`, individual.Student.Parent.Name, examType, individual.Student.Name, fmt.Sprintf("%d%s", individual.Student.Grade, individual.Student.GradeLabel))

		// Add the subject and score details
		for _, result := range individual.SubjectAndScoreResult {
			subjectName := result.Subject.Name
			score := "Belum Ada Nilai"
			if result.Score != nil {
				score = fmt.Sprintf("%.2f", *result.Score)
			}

			body += fmt.Sprintf("- Mata Pelajaran: %s | Nilai: %s\n", subjectName, score)
		}

		// Close the email with contact details
		body += fmt.Sprintf(`
		
		Jika terdapat pertanyaan atau memerlukan informasi lebih lanjut, Bapak/Ibu dapat menghubungi kami di %s.
		
		Terima kasih atas perhatian dan kerjasamanya.
		
		Hormat kami,
		Tim SINOAN`, m.schoolPhone)

		bodyTestScore = body
	} else {
		body := fmt.Sprintf(`
		SINOAN Service 🔔
		
		Kepada Yth. Bapak %s,
		
		Kami ingin memberikan informasi mengenai hasil ujian %s anak Anda, %s Kelas %s. Berikut adalah detail hasil ujian pada beberapa mata pelajaran:
		`, individual.Student.Parent.Name, examType, individual.Student.Name, fmt.Sprintf("%d%s", individual.Student.Grade, individual.Student.GradeLabel))

		// Add the subject and score details
		for _, result := range individual.SubjectAndScoreResult {
			subjectName := result.Subject.Name
			score := "Belum Ada Nilai"
			if result.Score != nil {
				score = fmt.Sprintf("%.2f", *result.Score)
			}

			body += fmt.Sprintf("- Mata Pelajaran: %s | Nilai: %s\n", subjectName, score)
		}

		// Close the email with contact details
		body += fmt.Sprintf(`
		
		Jika terdapat pertanyaan atau memerlukan informasi lebih lanjut, Bapak/Ibu dapat menghubungi kami di %s.
		
		Terima kasih atas perhatian dan kerjasamanya.
		
		Hormat kami,
		Tim SINOAN`, m.schoolPhone)

		bodyTestScore = body
	}

}

func (m *senderRepository) SendTestScores(ctx context.Context, examType string) error {
	var testScores []domain.TestScore
	var students []domain.Student
	var results []domain.IndividualExamScore
	// tNow := time.Now()
	// Fetch all test scores
	err := m.db.WithContext(ctx).
		Preload("Student").
		Preload("Subject").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id", "username", "name", "role", "created_at", "updated_at", "deleted_at")
		}).
		Where("deleted_at IS NULL").Find(&testScores).Error
	if err != nil {
		return err
	}

	// Fetch all students
	err = m.db.WithContext(ctx).Preload("Parent").Where("deleted_at IS NULL").Find(&students).Error
	if err != nil {
		return err
	}

	// Create a map for quick student lookup
	studentMap := make(map[int]domain.Student)
	for _, student := range students {
		studentMap[student.StudentID] = student
	}

	// Build results
	for _, ts := range testScores {
		// Find the student
		student, exists := studentMap[ts.StudentID]
		if !exists {
			continue // Skip if student not found (shouldn't happen in normal cases)
		}

		// Check if this student is already in the results
		var individual *domain.IndividualExamScore
		for i := range results {
			if results[i].StudentID == ts.StudentID {
				individual = &results[i]
				break
			}
		}

		// If the student doesn't exist, create a new entry
		if individual == nil {
			individual = &domain.IndividualExamScore{
				StudentID:             ts.StudentID,
				Student:               student,
				SubjectAndScoreResult: []domain.SubjectAndScoreResult{},
			}
			results = append(results, *individual)
		}

		// Add the subject and score to the student's results
		individual.SubjectAndScoreResult = append(individual.SubjectAndScoreResult, domain.SubjectAndScoreResult{
			SubjectID: ts.SubjectID,
			Subject:   ts.Subject,
			Score:     ts.Score,
		})
	}

	for _, idv := range results {
		var testScoreWAStatus, testScoreEmailStatus bool
		testScoreWAStatus, testScoreEmailStatus = false, false

		m.createTestScoreEmail(idv, examType)

		// send via mail
		if idv.Student.Parent.Email != nil && *idv.Student.Parent.Email != "" {
			if err := m.sendEmailTestScore(&idv); err != nil {
				fmt.Printf("Failed to send email to: %s\n", *idv.Student.Parent.Email)
				continue
			}
			testScoreEmailStatus = true
		}

		// send via wa
		err = m.sendWATestScore(ctx, &idv)
		if err != nil {
			testScoreWAStatus = false
		}
		testScoreWAStatus = true

		fmt.Println(testScoreWAStatus, testScoreEmailStatus)
	}

	return nil
}

func (m *senderRepository) SendMass(ctx context.Context, idList *[]int, userID *int, subjectID int) error {
	// Fetch the subject details
	var subject domain.Subject
	err := m.db.WithContext(ctx).Where("subject_id = ?", subjectID).First(&subject).Error
	if err != nil {
		return fmt.Errorf("failed to fetch subject details: %v", err)
	}

	for _, id := range *idList {
		var waStatus, emailStatus bool
		waStatus, emailStatus = false, false

		// Fetch student and parent details
		student, err := m.fetchStudentDetails(ctx, id)
		if err != nil {
			continue // Skip the current student if details cannot be fetched
		}

		// Initialize notification text with subject name
		err = m.initTextWithSubject(student, subject.Name)
		if err != nil {
			return err
		}

		// Attempt to send an email notification
		if student.Parent.Email != nil && *student.Parent.Email != "" {
			if err := m.sendEmail(student); err != nil {
				fmt.Printf("Failed to send email to: %s\n", *student.Parent.Email)
				continue
			}
			emailStatus = true
		}

		// Attempt to send a WhatsApp notification
		err = m.sendWA(ctx, student)
		if err != nil {
			fmt.Printf("Failed to send WhatsApp message to: %s\n", student.Parent.Telephone)
			continue
		}
		waStatus = true

		// Log the notification history
		err = m.logNotificationHistory(student.Student.StudentID, student.Student.ParentID, *userID, subjectID, waStatus, emailStatus)
		if err != nil {
			return fmt.Errorf("failed saving the data to notification history, error: %v", err)
		}
	}

	return nil
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

func (m *senderRepository) sendEmail(payload *domain.StudentAndParent) error {
	var msg string

	if payload.Parent.Gender == "female" {
		msg = "From: " + m.emailSender + "\n" +
			"To: " + *payload.Parent.Email + "\n" +
			"Subject: " + subject + "\n\n" +
			bodyFemaleEmail
	} else if payload.Parent.Gender == "male" {
		msg = "From: " + m.emailSender + "\n" +
			"To: " + *payload.Parent.Email + "\n" +
			"Subject: " + subject + "\n\n" +
			bodyMaleEmail
	}

	err := smtp.SendMail(m.smtpAdress, m.client, m.emailSender, []string{*payload.Parent.Email}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (m *senderRepository) sendEmailTestScore(idv *domain.IndividualExamScore) error {
	msgg := bodyTestScore
	err := smtp.SendMail(m.smtpAdress, m.client, m.emailSender, []string{*idv.Student.Parent.Email}, []byte(msgg))
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

func (m *senderRepository) sendWATestScore(ctx context.Context, idv *domain.IndividualExamScore) error {
	var msg string
	completeFormat := fmt.Sprintf("%s%s", "62", idv.Student.Parent.Telephone[1:])

	jid := types.NewJID(completeFormat, types.DefaultUserServer)

	msg = bodyTestScore

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

func (m *senderRepository) initTextWithSubject(payload *domain.StudentAndParent, subjectName string) error {
	tNow := time.Now()

	// Format the date and time
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

	bodyMale = fmt.Sprintf(`
SINOAN Service 🔔

Kepada Yth. Bapak %s,

Kami ingin memberitahukan bahwa anak Bapak, %s, tidak hadir di pelajaran "%s" pada tanggal %s pukul %s %s.

Alasan ketidakhadiran belum kami terima hingga saat ini. Kami berharap Bapak dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi anak Bapak.

Jika terdapat pertanyaan atau memerlukan bantuan lebih lanjut, Bapak dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, payload.Parent.Name, payload.Student.Name, strings.ToUpper(subjectName), formattedDate, hourAndMinute, isAM, m.schoolPhone)

	bodyFemale = fmt.Sprintf(`
SINOAN Service 🔔

Kepada Yth. Ibu %s,

Kami ingin memberitahukan bahwa anak Ibu, %s, tidak hadir di pelajaran "%s" pada tanggal %s pukul %s %s.

Alasan ketidakhadiran belum kami terima hingga saat ini. Kami berharap Ibu dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi anak Ibu.

Jika terdapat pertanyaan atau memerlukan bantuan lebih lanjut, Ibu dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, payload.Parent.Name, payload.Student.Name, strings.ToUpper(subjectName), formattedDate, hourAndMinute, isAM, m.schoolPhone)

	// email

	bodyMaleEmail = fmt.Sprintf(`
SINOAN Service 🔔

Kepada Yth. Bapak %s,

Kami ingin memberitahukan bahwa anak Bapak, %s, tidak hadir di pelajaran "%s" pada tanggal %s pukul %s %s.

Alasan ketidakhadiran belum kami terima hingga saat ini. Kami berharap Bapak dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi anak Bapak.

Jika terdapat pertanyaan atau memerlukan bantuan lebih lanjut, Bapak dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, payload.Parent.Name, payload.Student.Name, strings.ToUpper(subjectName), formattedDate, hourAndMinute, isAM, m.schoolPhone)

	bodyFemaleEmail = fmt.Sprintf(`
SINOAN Service 🔔

Kepada Yth. Ibu %s,

Kami ingin memberitahukan bahwa anak Ibu, %s, tidak hadir di pelajaran "%s" pada tanggal %s pukul %s %s.

Alasan ketidakhadiran belum kami terima hingga saat ini. Kami berharap Ibu dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi anak Ibu.

Jika terdapat pertanyaan atau memerlukan bantuan lebih lanjut, Ibu dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, payload.Parent.Name, payload.Student.Name, strings.ToUpper(subjectName), formattedDate, hourAndMinute, isAM, m.schoolPhone)

	return nil
}

func (m *senderRepository) logNotificationHistory(studentID, parentID, userID, subjectID int, whatsappSuccess, emailSuccess bool) error {
	history := &domain.AttendanceNotificationHistory{
		StudentID:      studentID,
		ParentID:       parentID,
		UserID:         userID,
		SubjectID:      subjectID,
		WhatsappStatus: whatsappSuccess,
		EmailStatus:    emailSuccess,
	}

	err := m.db.Create(history).Error
	if err != nil {
		return fmt.Errorf("could not log notification history: %v", err)
	}

	return nil
}
