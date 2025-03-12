package amqprpc

import (
	"tourism-backend/internal/usecase"
	"tourism-backend/pkg/rabbitmq/rmq_rpc/server"
)

// NewRouter -.
func NewRouter(t usecase.TourismInterface) map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)
	{
		newTourismRoutes(routes, t)
	}

	return routes
}
