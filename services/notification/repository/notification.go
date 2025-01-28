package repository

import (
	"context"
	"fmt"
	"notification/domain"

	"gorm.io/gorm"
)

type notificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) domain.NotificationRepo {
	return &notificationRepo{
		db: db,
	}
}

func (np *notificationRepo) GetAllAttendanceNotificationHistory(ctx context.Context) (*[]domain.AttendanceNotificationHistoryResponse, error) {
	var dataHolder []domain.AttendanceNotificationHistory
	var finalDatas []domain.AttendanceNotificationHistoryResponse

	// Query the attendance notification history
	if err := np.db.WithContext(ctx).
		Preload("Student").
		Preload("Parent").
		Preload("User").
		Preload("Subject").
		Find(&dataHolder).Error; err != nil {
		return nil, fmt.Errorf("could not get all attendance notification history, error: %v", err)
	}

	// Iterate over the fetched records to prepare the response
	for _, record := range dataHolder {
		if record.Student.StudentID == 0 || record.Subject.SubjectID == 0 {
			continue
		}

		userResponse := domain.UserResponse{
			UserID:    record.User.UserID,
			Username:  record.User.Username,
			Name:      record.User.Name,
			Role:      record.User.Role,
			CreatedAt: record.User.CreatedAt,
			UpdatedAt: record.User.UpdatedAt,
			DeletedAt: record.User.DeletedAt,
		}

		// Append to final response slice
		finalDatas = append(finalDatas, domain.AttendanceNotificationHistoryResponse{
			Student:        record.Student,
			Parent:         record.Parent,
			User:           userResponse,
			Subject:        record.Subject,
			WhatsappStatus: record.WhatsappStatus,
			EmailStatus:    record.EmailStatus,
			CreatedAt:      record.CreatedAt,
		})
	}

	return &finalDatas, nil
}
