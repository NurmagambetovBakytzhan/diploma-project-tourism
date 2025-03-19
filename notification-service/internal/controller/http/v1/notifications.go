package v1

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"log"
	"notification-service/internal/usecase"
	"notification-service/pkg/logger"
	pkg "notification-service/pkg/websocket"
	"notification-service/utils"
)

type notificationRoutes struct {
	l logger.Interface
}

func newTourismRoutes(handler fiber.Router, t *usecase.NotificationsUseCase, l logger.Interface) {
	r := &notificationRoutes{l}
	wshandler := handler.Group("/ws")
	wshandler.Use(utils.JWTAuthMiddleware(), utils.WebSocketMiddleware())
	{
		wshandler.Get("/", websocket.New(r.WebSocketHandler))
	}
}

func (r *notificationRoutes) WebSocketHandler(c *websocket.Conn) {
	clientObj := pkg.ClientObject{
		UserID: c.Locals("userID").(string),
		Conn:   c,
	}
	defer func() {
		pkg.Unregister <- clientObj
		c.Close()
	}()

	// Register the client
	pkg.Register <- clientObj
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}
			return // Calls the deferred function, i.e. closes the connection on error
		}
	}
}
