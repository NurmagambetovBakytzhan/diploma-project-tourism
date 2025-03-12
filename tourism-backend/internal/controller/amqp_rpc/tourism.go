package amqprpc

import (
	"tourism-backend/internal/usecase"
	"tourism-backend/pkg/rabbitmq/rmq_rpc/server"
)

type tourismRoutes struct {
	tourismUseCase usecase.TourismInterface
}

func newTourismRoutes(routes map[string]server.CallHandler, t usecase.TourismInterface) {
	//r := &tourismRoutes{t}
	//{
	//	// routes["getHistory"] = r.getHistory()
	//}
}

// type historyResponse struct {
// 	History []entity.Tour `json:"history"`
// }

// func (r *tourismRoutes) getHistory() server.CallHandler {
// 	return func(d *amqp.Delivery) (interface{}, error) {
// 		tourisms, err := r.tourismUseCase.History(context.Background())
// 		if err != nil {
// 			return nil, fmt.Errorf("amqp_rpc - tourismRoutes - getHistory - r.tourismUseCase.History: %w", err)
// 		}

// 		response := historyResponse{tourisms}

// 		return response, nil
// 	}
// }
