package repo

import (
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
