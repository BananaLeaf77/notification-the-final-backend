package repository

import (
	"context"
	"fmt"
	"net/smtp"
	"notification/domain"
	"os"
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
`, individual.Student.Parent.Name, examType, individual.Student.StudentNSN, individual.Student.Name, fmt.Sprintf("%d %s", individual.Student.Grade, individual.Student.GradeLabel))

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
`, individual.Student.Parent.Name, examType, individual.Student.StudentNSN, individual.Student.Name, fmt.Sprintf("%d %s", individual.Student.Grade, individual.Student.GradeLabel))

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

func (m *senderRepository) buatNilaiTesEmail(individual domain.IndividualExamScore, examType string) string {
	// Mulai dengan pengantar umum
	if individual.Student.Parent.Gender != "male" {
		body := fmt.Sprintf(`Layanan SINOAN 🔔

Yth. Ibu %s,
Kami ingin memberitahukan tentang hasil %s untuk siswa berikut:
NSN: %s,
Nama: %s,
Kelas: %s.
Berikut adalah detail hasil ujian untuk beberapa mata pelajaran:
`, individual.Student.Parent.Name, examType, individual.Student.StudentNSN, individual.Student.Name, fmt.Sprintf("%d %s", individual.Student.Grade, individual.Student.GradeLabel))

		// Tambahkan detail mata pelajaran dan nilai
		for _, result := range individual.SubjectAndScoreResult {
			subjectName := result.Subject.Name
			score := "Belum Ada Nilai | 0"
			if result.Score != nil {
				score = fmt.Sprintf("%.1f", *result.Score)
			}

			body += fmt.Sprintf("- Kode (%s) | Mata Pelajaran: %s | Nilai: %s\n", result.Subject.SubjectCode, subjectName, score)
		}

		// Tutup email dengan detail kontak
		body += fmt.Sprintf(`
Jika ibu memiliki pertanyaan atau membutuhkan informasi lebih lanjut, ibu dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.

Hormat kami,
Tim SINOAN`, m.schoolPhone)

		bodyTestScore := body
		return bodyTestScore
	} else {
		body := fmt.Sprintf(`Layanan SINOAN 🔔

Yth. Bapak %s,
Kami ingin memberitahukan tentang hasil %s untuk siswa berikut:
NSN: %s,
Nama: %s,
Kelas: %s.
Berikut adalah detail hasil ujian untuk beberapa mata pelajaran:
`, individual.Student.Parent.Name, examType, individual.Student.StudentNSN, individual.Student.Name, fmt.Sprintf("%d %s", individual.Student.Grade, individual.Student.GradeLabel))

		// Tambahkan detail mata pelajaran dan nilai
		for _, result := range individual.SubjectAndScoreResult {
			subjectName := result.Subject.Name
			score := "Belum Ada Nilai | 0"
			if result.Score != nil {
				score = fmt.Sprintf("%.1f", *result.Score)
			}

			body += fmt.Sprintf("- Kode (%s) | Mata Pelajaran: %s | Nilai: %s\n", result.Subject.SubjectCode, subjectName, score)
		}

		// Tutup email dengan detail kontak
		body += fmt.Sprintf(`
Jika bapak memiliki pertanyaan atau membutuhkan informasi lebih lanjut, bapak dapat menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.

Hormat kami,
Tim SINOAN`, m.schoolPhone)

		bodyTestScore := body
		return bodyTestScore
	}
}

func (m *senderRepository) SendTestScores(ctx context.Context, examType string) error {
	var testScores []domain.TestScore
	var students []domain.Student
	var resultsMap = make(map[string]domain.IndividualExamScore)
	langValue := os.Getenv("MESSENGER_LANGUAGE")
	langValueLowered := strings.ToLower(langValue)
	var examTypeProcessed string
	fmt.Println(examType)

	// Process exam type based on language
	if langValueLowered == "ind" {
		switch examType {
		case "Midterm Tests":
			examTypeProcessed = "Ulangan Tengah Semester (UTS)"
		case "End of Semester Tests":
			examTypeProcessed = "Ulangan Akhir Semester (UAS)"
		default:
			examTypeProcessed = examType
		}
	} else {
		examTypeProcessed = examType
	}

	fmt.Println(examTypeProcessed)

	// Fetch all test scores with related data
	err := m.db.WithContext(ctx).
		Preload("Student").
		Preload("Subject").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id", "username", "name", "role", "created_at", "updated_at", "deleted_at")
		}).
		Where("sent_at IS NULL").
		Find(&testScores).Error
	if err != nil {
		return fmt.Errorf("failed to fetch test scores: %w", err)
	}

	if len(testScores) == 0 {
		return fmt.Errorf("theres no any test scores to be sent")
	}

	// Extract student IDs from test scores
	studentIDs := make([]string, 0, len(testScores))
	for _, score := range testScores {
		studentIDs = append(studentIDs, score.StudentNSN)
	}

	// Fetch all students associated with the test scores
	err = m.db.WithContext(ctx).
		Preload("Parent").
		Where("student_nsn IN (?)", studentIDs).
		Find(&students).Error
	if err != nil {
		return fmt.Errorf("failed to fetch students: %w", err)
	}

	// Build a map of students for quick lookup
	studentMap := make(map[string]domain.Student, len(students))
	for _, student := range students {
		studentMap[student.StudentNSN] = student
	}

	// Build results map
	for _, score := range testScores {
		student, exists := studentMap[score.StudentNSN]
		if !exists {
			continue
		}

		individual, found := resultsMap[student.StudentNSN]
		if !found {
			individual = domain.IndividualExamScore{
				StudentNSN:            student.StudentNSN,
				Student:               student,
				SubjectAndScoreResult: []domain.SubjectAndScoreResult{},
			}
		}

		// Prevent duplicate subjects
		duplicate := false
		for _, subject := range individual.SubjectAndScoreResult {
			if subject.SubjectCode == score.SubjectCode {
				duplicate = true
				break
			}
		}
		if !duplicate {
			individual.SubjectAndScoreResult = append(individual.SubjectAndScoreResult, domain.SubjectAndScoreResult{
				SubjectCode: score.SubjectCode,
				Subject:     score.Subject,
				Score:       score.Score,
			})
		}

		resultsMap[student.StudentNSN] = individual
	}

	// Convert the resultsMap to a slice
	results := make([]domain.IndividualExamScore, 0, len(resultsMap))
	for _, result := range resultsMap {
		results = append(results, result)
	}

	// Worker pool to limit concurrency
	const maxWorkers = 10 // Adjust based on system capacity
	var wg sync.WaitGroup
	workerPool := make(chan struct{}, maxWorkers)
	errChan := make(chan error, len(results)) // Channel to collect errors

	// Process results concurrently
	for _, idv := range results {
		wg.Add(1)
		workerPool <- struct{}{} // Acquire a worker slot

		go func(idv domain.IndividualExamScore) {
			defer wg.Done()
			defer func() { <-workerPool }() // Release the worker slot

			var messageString string
			if langValueLowered == "ind" {
				messageString = m.buatNilaiTesEmail(idv, examTypeProcessed)
			} else {
				messageString = m.createTestScoreEmail(idv, examTypeProcessed)
			}

			// Send email
			if idv.Student.Parent.Email != nil && *idv.Student.Parent.Email != "" {
				if err := m.sendEmailTestScore(&idv, messageString); err != nil {
					errChan <- fmt.Errorf("failed to send email to %s: %w", *idv.Student.Parent.Email, err)
					return
				}
			}

			// Send WhatsApp
			if err := m.sendWATestScore(ctx, &idv, messageString); err != nil {
				errChan <- fmt.Errorf("failed to send WhatsApp for student %s: %w", idv.StudentNSN, err)
				return
			}
		}(idv)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan) // Close the error channel

	// Collect errors from the error channel
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// Log all errors
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println("Error:", err)
		}
		return fmt.Errorf("encountered %d errors while sending test scores", len(errors))
	}

	// Mark test scores as deleted
	err = m.db.WithContext(ctx).
		Model(&domain.TestScore{}).
		Where("sent_at IS NULL").
		Updates(map[string]interface{}{
			"sent_at": time.Now(),
			"type":    examTypeProcessed,
		}).Error

	if err != nil {
		return fmt.Errorf("failed to soft delete all test scores: %w", err)
	}

	return nil
}

func (m *senderRepository) SendMass(ctx context.Context, nsnList *[]string, userID *int, subjectCode string) error {
	// Fetch the subject details
	langValue := os.Getenv("MESSENGER_LANGUAGE")
	langValueLowered := strings.ToLower(langValue)
	var subject domain.Subject
	err := m.db.WithContext(ctx).Where("subject_code = ?", subjectCode).First(&subject).Error
	if err != nil {
		return fmt.Errorf("failed to fetch subject details: %v", err)
	}

	for _, nsn := range *nsnList {
		var waStatus, emailStatus bool
		waStatus, emailStatus = false, false

		// Fetch student and parent details
		student, err := m.fetchStudentDetails(ctx, nsn)
		if err != nil {
			continue // Skip the current student if details cannot be fetched
		}

		var subjectForEmailSender *string
		var body *string
		if langValueLowered == "ind" {
			// Initialize notification text with subject name
			subjectForEmailSender, body, err = m.inisialisasiTeksDenganSubjek(student, subject.Name)
			if err != nil {
				return err
			}
		} else {
			// Initialize notification text with subject name
			subjectForEmailSender, body, err = m.initTextWithSubject(student, subject.Name)
			if err != nil {
				return err
			}
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
		err = m.logNotificationHistory(student.Student.StudentNSN, subjectCode, student.Student.ParentID, *userID, waStatus, emailStatus)
		if err != nil {
			return fmt.Errorf("failed saving the data to notification history, error: %v", err)
		}
	}

	return nil
}

func (m *senderRepository) fetchStudentDetails(ctx context.Context, nsn string) (*domain.StudentAndParent, error) {
	var student domain.Student
	var parent domain.Parent

	err := m.db.WithContext(ctx).Where("student_nsn = ?", nsn).Preload("Parent").First(&student).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("student with StudentNSN %s not found", nsn)
		}
		return nil, fmt.Errorf("could not fetch student details: %v", err)
	}

	err = m.db.WithContext(ctx).Where("parent_id = ? AND deleted_at IS NULL", student.ParentID).First(&parent).Error
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
		bodyMale := fmt.Sprintf(`SINOAN Service 🔔

Dear Mr. %s,

We would like to inform you that your child,

NSN: %s,
Name: %s, 
Class: %d %s.

was absent from the lesson "%s" on %s at %s %s.

We have not yet received any reason for the absence. We kindly ask you to provide confirmation or further information regarding your child's condition.

If you have any questions or require further assistance, please feel free to contact us at %s.

Thank you for your attention and cooperation.`, payload.Parent.Name, payload.Student.StudentNSN, payload.Student.Name, payload.Student.Grade, payload.Student.GradeLabel, strings.ToUpper(subjectName), formattedDate, hourAndMinute, isAM, m.schoolPhone)

		return &subject, &bodyMale, nil
	} else {
		bodyFemale := fmt.Sprintf(`SINOAN Service 🔔

Dear Mrs. %s,

We would like to inform you that your child, 

NSN: %s,
Name: %s, 
Class: %d %s.

was absent from the lesson "%s" on %s at %s %s.

We have not yet received any reason for the absence. We kindly ask you to provide confirmation or further information regarding your child's condition.

If you have any questions or require further assistance, please feel free to contact us at %s.

Thank you for your attention and cooperation.`, payload.Parent.Name, payload.Student.StudentNSN, payload.Student.Name, payload.Student.Grade, payload.Student.GradeLabel, strings.ToUpper(subjectName), formattedDate, hourAndMinute, isAM, m.schoolPhone)
		return &subject, &bodyFemale, nil
	}
}

func (m *senderRepository) inisialisasiTeksDenganSubjek(payload *domain.StudentAndParent, subjectName string) (*string, *string, error) {
	tNow := time.Now()

	// Format tanggal dan waktu
	formattedDate := tNow.Format("02/01/2006") // Format DD/MM/YYYY
	hourOnly := tNow.Format("15")              // Format 24 jam
	hourAndMinute := tNow.Format("15:04")      // Format HH:MM

	intHourOnly, err := strconv.Atoi(hourOnly)
	if err != nil {
		return nil, nil, err
	}

	isAM := "AM"
	if intHourOnly >= 12 {
		isAM = "PM"
	}

	subject := fmt.Sprintf("Pemberitahuan Ketidakhadiran untuk %s pada %s %s tanggal %s", payload.Student.Name, hourAndMinute, isAM, formattedDate)

	if payload.Parent.Gender == "male" {
		bodyMale := fmt.Sprintf(`Layanan SINOAN 🔔

Yth. Bapak %s,

Kami ingin memberitahukan bahwa anak bapak,

NSN: %s,
Nama: %s, 
Kelas: %d %s.

tidak hadir pada pelajaran "%s" tanggal %s pukul %s %s.

Kami belum menerima alasan ketidakhadiran tersebut. Kami mohon bapak dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi anak bapak.

Jika bapak memiliki pertanyaan atau membutuhkan bantuan lebih lanjut, jangan ragu untuk menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, payload.Parent.Name, payload.Student.StudentNSN, payload.Student.Name, payload.Student.Grade, payload.Student.GradeLabel, strings.ToUpper(subjectName), formattedDate, hourAndMinute, isAM, m.schoolPhone)

		return &subject, &bodyMale, nil
	} else {
		bodyFemale := fmt.Sprintf(`Layanan SINOAN 🔔

Yth. Ibu %s,

Kami ingin memberitahukan bahwa anak ibu, 

NSN: %s,
Nama: %s, 
Kelas: %d %s.

tidak hadir pada pelajaran "%s" tanggal %s pukul %s %s.

Kami belum menerima alasan ketidakhadiran tersebut. Kami mohon ibu dapat memberikan konfirmasi atau informasi lebih lanjut mengenai kondisi anak ibu.

Jika ibu memiliki pertanyaan atau membutuhkan bantuan lebih lanjut, jangan ragu untuk menghubungi kami di %s.

Terima kasih atas perhatian dan kerjasamanya.`, payload.Parent.Name, payload.Student.StudentNSN, payload.Student.Name, payload.Student.Grade, payload.Student.GradeLabel, strings.ToUpper(subjectName), formattedDate, hourAndMinute, isAM, m.schoolPhone)
		return &subject, &bodyFemale, nil
	}
}

func (m *senderRepository) logNotificationHistory(StudentNSN, subjectCode string, parentID, userID int, whatsappSuccess, emailSuccess bool) error {
	history := &domain.AttendanceNotificationHistory{
		StudentNSN:     StudentNSN,
		ParentID:       parentID,
		UserID:         userID,
		SubjectCode:    subjectCode,
		WhatsappStatus: whatsappSuccess,
		EmailStatus:    emailSuccess,
	}

	err := m.db.Create(history).Error
	if err != nil {
		return fmt.Errorf("could not log notification history: %v", err)
	}

	return nil
}
