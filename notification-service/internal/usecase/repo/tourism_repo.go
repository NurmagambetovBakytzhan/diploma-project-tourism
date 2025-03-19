package repo

import (
	"notification-service/pkg/postgres"
)

const _defaultEntityCap = 64

// NotificationsRepo -.
type NotificationsRepo struct {
	PG *postgres.Postgres
}

// New -.
func NewNotificationsRepo(pg *postgres.Postgres) *NotificationsRepo {
	return &NotificationsRepo{pg}
}
