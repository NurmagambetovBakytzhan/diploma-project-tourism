package repo

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"notification-service/internal/entity"
	"notification-service/pkg/postgres"
)

type NotificationRepo struct {
	PG *postgres.Postgres
}

// New -.
func NewNotificationRepo(pg *postgres.Postgres) *NotificationRepo {
	return &NotificationRepo{pg}
}

func (u *NotificationRepo) CreateNotification(notif *entity.Notification) error {
	err := u.PG.Conn.Create(&notif).Error
	if err != nil {
		log.Println("Error CreateNotification: ", err)
		return err
	}
	return nil
}
func (n *NotificationRepo) GetMyNotifications(userID uuid.UUID) ([]*entity.Notification, error) {
	var notifications []*entity.Notification
	err := n.PG.Conn.Table("notification_service.notifications").
		Where("recipient_id = ?", userID).
		Find(&notifications).Error
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("failed to fetch notifications")
	}
	return notifications, nil
}
