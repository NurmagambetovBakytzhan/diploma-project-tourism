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

func newNotificationRoutes(handler fiber.Router, t *usecase.NotificationsUseCase, l logger.Interface) {
	r := &notificationRoutes{l}
	wshandler := handler.Group("/notifications/ws")
	wshandler.Use(utils.JWTAuthMiddleware(), utils.WebSocketMiddleware())
	{
		wshandler.Get("/", websocket.New(r.WebSocketHandler))
	}
}

// WebSocketHandler @Summary WebSocket Connection for Notifications
// @Description Establishes a WebSocket connection to receive real-time notifications.
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /v1/notifications/ws [get]
// @Security Bearer
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
