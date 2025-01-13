package usecase

import (
	"context"
	"notification/domain"
)

type notificationUC struct {
	repo domain.NotificationRepo
}

func NewNotificationUseCase(repo domain.NotificationRepo) domain.NotificationUseCase {
	return &notificationUC{
		repo: repo,
	}
}

func (nuc *notificationUC) GetAllAttendanceNotificationHistory(ctx context.Context) (*[]domain.AttendanceNotificationHistoryResponse, error) {
	datas, err := nuc.repo.GetAllAttendanceNotificationHistory(ctx)
	if err != nil {
		return nil, err
	}
	return datas, nil
}
