package domain

import "time"

type NotificationHistory struct {
	NotificationHistoryID int       `gorm:"primaryKey;autoIncrement" json:"notification_history_id"`
	StudentID             int       `gorm:"not null" json:"student_id"`
	ParentID              int       `gorm:"not null" json:"parent_id"`
	UserID                int       `gorm:"not null" json:"user_id"`
	WhatsappStatus        bool      `gorm:"not null" json:"whatsapp"`
	EmailStatus           bool      `gorm:"not null" json:"email"`
	CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`
}
