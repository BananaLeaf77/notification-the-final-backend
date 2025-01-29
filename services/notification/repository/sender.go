package repository

import (
	"context"
	"fmt"
	"net/smtp"
	"notification/domain"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

// init var
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

func (m *senderRepository) createTestScoreEmail(individual domain.IndividualExamScore, examType string) string {
	// Start with a general introduction
	if individual.Student.Parent.Gender != "male" {
		body := fmt.Sprintf(`SINOAN Service 🔔

Dear Mrs. %s,
We would like to inform you about the %s results for the following student:
NSN: %s,
Name: %s,
Class: %s.
Below are the details of the test results for several subjects:
`, individual.Student.Parent.Name, examType, individual.Student.NSN, individual.Student.Name, fmt.Sprintf("%d%s", individual.Student.Grade, individual.Student.GradeLabel))

		// Add the subject and score details
		for _, result := range individual.SubjectAndScoreResult {
			subjectName := result.Subject.Name
			score := "No Score Yet | 0"
			if result.Score != nil {
				score = fmt.Sprintf("%.1f", *result.Score)
			}

			body += fmt.Sprintf("- Code (%s) | Subject: %s | Score: %s\n", result.Subject.SubjectCode, subjectName, score)
		}

		// Close the email with contact details
		body += fmt.Sprintf(`

If you have any questions or need further information, you can contact us at %s.

Thank you for your attention and cooperation.

Sincerely,
SINOAN Team`, m.schoolPhone)

		bodyTestScore := body
		return bodyTestScore
	} else {
		body := fmt.Sprintf(`SINOAN Service 🔔

Dear Mr. %s,
We would like to inform you about the %s results for the following student:
NSN: %s,
Name: %s,
Class: %s.
Below are the details of the test results for several subjects:
`, individual.Student.Parent.Name, examType, individual.Student.NSN, individual.Student.Name, fmt.Sprintf("%d%s", individual.Student.Grade, individual.Student.GradeLabel))

		// Add the subject and score details
		for _, result := range individual.SubjectAndScoreResult {
			subjectName := result.Subject.Name
			score := "No Score Yet | 0"
			if result.Score != nil {
				score = fmt.Sprintf("%.1f", *result.Score)
			}

			body += fmt.Sprintf("- Code (%s) | Subject: %s | Score: %s\n", result.Subject.SubjectCode, subjectName, score)
		}

		// Close the email with contact details
		body += fmt.Sprintf(`

If you have any questions or need further information, you can contact us at %s.

Thank you for your attention and cooperation.

Sincerely,
SINOAN Team`, m.schoolPhone)

		bodyTestScore := body
		return bodyTestScore
	}

}

func (m *senderRepository) SendTestScores(ctx context.Context, examType string) error {
	var testScores []domain.TestScore
	var students []domain.Student
	var results []domain.IndividualExamScore
	studentMap := make(map[int]domain.Student)
	tNow := time.Now()

	// Fetch all test scores with related data
	err := m.db.WithContext(ctx).
		Preload("Student").
		Preload("Subject").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id", "username", "name", "role", "created_at", "updated_at", "deleted_at")
		}).
		Where("deleted_at IS NULL").
		Find(&testScores).Error
	if err != nil {
		return fmt.Errorf("failed to fetch test scores: %w", err)
	}

	// Extract student IDs from test scores
	studentIDs := make([]int, 0, len(testScores))
	for _, score := range testScores {
		studentIDs = append(studentIDs, score.StudentID)
	}

	// Fetch all students associated with the test scores
	err = m.db.WithContext(ctx).
		Preload("Parent").
		Where("student_id IN (?)", studentIDs).
		Find(&students).Error
	if err != nil {
		return fmt.Errorf("failed to fetch students: %w", err)
	}

	// Build a map of students for quick lookup
	for _, student := range students {
		studentMap[student.StudentID] = student
	}

	// Build results for each test score
	for _, ts := range testScores {
		student, exists := studentMap[ts.StudentID]
		if !exists {
			continue // Skip if student not found
		}

		// Find or create an individual exam score entry for the student
		var individual *domain.IndividualExamScore
		for i := range results {
			if results[i].StudentID == ts.StudentID {
				individual = &results[i]
				break
			}
		}

		// Create a new entry if one doesn't exist
		if individual == nil {
			newEntry := domain.IndividualExamScore{
				StudentID:             ts.StudentID,
				Student:               student,
				SubjectAndScoreResult: []domain.SubjectAndScoreResult{},
			}
			results = append(results, newEntry)
			individual = &results[len(results)-1]
		}

		// Add the subject and score to the student's results
		individual.SubjectAndScoreResult = append(individual.SubjectAndScoreResult, domain.SubjectAndScoreResult{
			SubjectID: ts.SubjectID,
			Subject:   ts.Subject,
			Score:     ts.Score,
		})
	}

	// Use WaitGroup to manage Go routines
	var wg sync.WaitGroup
	errorChan := make(chan error, len(results)*2) // Buffer size for errors

	// Process results concurrently
	for _, idv := range results {
		wg.Add(2) // Two goroutines: one for email, one for WhatsApp
		strBody := m.createTestScoreEmail(idv, examType)

		// Send email
		go func(idv domain.IndividualExamScore) {
			defer wg.Done()
			if idv.Student.Parent.Email != nil && *idv.Student.Parent.Email != "" {
				if err := m.sendEmailTestScore(&idv, strBody); err != nil {
					errorChan <- fmt.Errorf("failed to send email to: %s, error: %w", *idv.Student.Parent.Email, err)
				}
			}
		}(idv)

		// Send WhatsApp
		go func(idv domain.IndividualExamScore) {
			defer wg.Done()
			if err := m.sendWATestScore(ctx, &idv, strBody); err != nil {
				errorChan <- fmt.Errorf("failed to send WhatsApp for student ID: %d, error: %w", idv.StudentID, err)
			}
		}(idv)
	}

	// Wait for all Go routines to complete
	wg.Wait()
	close(errorChan)

	// Collect and log errors
	for err := range errorChan {
		fmt.Println("Error:", err)
	}

	// Mark test scores as deleted
	for i := range testScores {
		testScores[i].DeletedAt = &tNow
	}

	// Batch update in the database
	err = m.db.WithContext(ctx).Save(&testScores).Error
	if err != nil {
		return fmt.Errorf("failed to soft delete test scores: %w", err)
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
		subjectForEmailSender, body, err := m.initTextWithSubject(student, subject.Name)
		if err != nil {
			return err
		}

		// Attempt to send an email notification
		if student.Parent.Email != nil && *student.Parent.Email != "" {
			if err := m.sendEmail(student, *subjectForEmailSender, *body); err != nil {
				fmt.Printf("Failed to send email to: %s\n", *student.Parent.Email)
				continue
			}
			emailStatus = true
		}

		// Attempt to send a WhatsApp notification
		err = m.sendWA(ctx, student, *body)
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

func (m *senderRepository) sendEmail(payload *domain.StudentAndParent, subjectEmail string, body string) error {
	msg := "From: " + m.emailSender + "\r\n" +
		"To: " + *payload.Parent.Email + "\r\n" +
		"Subject: " + subjectEmail + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	err := smtp.SendMail(m.smtpAdress, m.client, m.emailSender, []string{*payload.Parent.Email}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (m *senderRepository) sendEmailTestScore(idv *domain.IndividualExamScore, body string) error {

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

	subjectTestScoreEmail := fmt.Sprintf("Pemberitahuan Hasil Penilaian %s pada %s %s, tanggal %s", idv.Student.Name, hourAndMinute, isAM, formattedDate)

	msg := "From: " + m.emailSender + "\r\n" +
		"To: " + *idv.Student.Parent.Email + "\r\n" +
		"Subject: " + subjectTestScoreEmail + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	err = smtp.SendMail(m.smtpAdress, m.client, m.emailSender, []string{*idv.Student.Parent.Email}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (m *senderRepository) sendWA(ctx context.Context, payload *domain.StudentAndParent, body string) error {
	completeFormat := fmt.Sprintf("%s%s", "62", payload.Parent.Telephone[1:])

	jid := types.NewJID(completeFormat, types.DefaultUserServer)

	conversationMessage := &waE2E.Message{
		Conversation: &body,
	}

	_, err := m.meowClient.SendMessage(ctx, jid, conversationMessage)
	if err != nil {
		fmt.Println("meow client error")
		return err
	}
	return nil
}

func (m *senderRepository) sendWATestScore(ctx context.Context, idv *domain.IndividualExamScore, strBody string) error {
	completeFormat := fmt.Sprintf("%s%s", "62", idv.Student.Parent.Telephone[1:])

	jid := types.NewJID(completeFormat, types.DefaultUserServer)

	conversationMessage := &waE2E.Message{
		Conversation: &strBody,
	}

	_, err := m.meowClient.SendMessage(ctx, jid, conversationMessage)
	if err != nil {
		return err
	}
	return nil
}

func (m *senderRepository) initTextWithSubject(payload *domain.StudentAndParent, subjectName string) (*string, *string, error) {
	tNow := time.Now()

	// Format the date and time
	formattedDate := tNow.Format("02/01/2006") // DD/MM/YYYY format
	hourOnly := tNow.Format("15")              // 24-hour format
	hourAndMinute := tNow.Format("15:04")      // HH:MM format

	intHourOnly, err := strconv.Atoi(hourOnly)
	if err != nil {
		return nil, nil, err
	}

	isAM := "AM"
	if intHourOnly >= 12 {
		isAM = "PM"
	}

	subject := fmt.Sprintf("Notification of Absence for %s at %s %s on %s", payload.Student.Name, hourAndMinute, isAM, formattedDate)

	if payload.Parent.Gender == "male" {
		bodyMale := fmt.Sprintf(`
SINOAN Service 🔔

Dear Mr. %s,

We would like to inform you that your child,

NSN: %s,
Name: %s, 
class %d%s 

was absent from the lesson "%s" on %s at %s %s.

We have not yet received any reason for the absence. We kindly ask you to provide confirmation or further information regarding your child's condition.

If you have any questions or require further assistance, please feel free to contact us at %s.

Thank you for your attention and cooperation.`, payload.Parent.Name, payload.Student.NSN, payload.Student.Name, payload.Student.Grade, payload.Student.GradeLabel, strings.ToUpper(subjectName), formattedDate, hourAndMinute, isAM, m.schoolPhone)

		return &subject, &bodyMale, nil
	} else {
		bodyFemale := fmt.Sprintf(`
SINOAN Service 🔔

Dear Mrs. %s,

We would like to inform you that your child, 

NSN: %s,
Name: %s, 
class %d%s 

was absent from the lesson "%s" on %s at %s %s.

We have not yet received any reason for the absence. We kindly ask you to provide confirmation or further information regarding your child's condition.

If you have any questions or require further assistance, please feel free to contact us at %s.

Thank you for your attention and cooperation.`, payload.Parent.Name, payload.Student.NSN, payload.Student.Name, payload.Student.Grade, payload.Student.GradeLabel, strings.ToUpper(subjectName), formattedDate, hourAndMinute, isAM, m.schoolPhone)
		return &subject, &bodyFemale, nil
	}
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
