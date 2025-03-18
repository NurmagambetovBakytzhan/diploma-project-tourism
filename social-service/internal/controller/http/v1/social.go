package v1

import (
	"github.com/casbin/casbin/v2"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"log"
	"social-service/internal/entity"
	"social-service/internal/usecase"
	"social-service/pkg/logger"
	pkg "social-service/pkg/websocket"
	"social-service/utils"
)

type socialRoutes struct {
	l logger.Interface
	s usecase.SocialInterface
}

func newSocialRoutes(handler fiber.Router, l logger.Interface, csbn *casbin.Enforcer, s *usecase.SocialUseCase) {
	r := &socialRoutes{l, s}

	h := handler.Group("/social")
	h.Use(utils.JWTAuthMiddleware())
	{
		h.Post("/chat", r.CreateChat)
		h.Post("/chat/enter", r.EnterToChat)
		wshandler := handler.Group("/ws")
		wshandler.Use(utils.WebSocketMiddleware())
		{
			wshandler.Get("/", websocket.New(r.WebSocketHandler))
		}
	}
}

//
//func (r *socialRoutes) GetAllChats(c *fiber.Ctx) error {
//	perPage := c.Query("per_page", "10")
//	sortOrder := c.Query("sort_order", "desc")
//	cursor := c.Query("cursor", "")
//	limit, err := strconv.ParseInt(perPage, 10, 64)
//	if limit < 1 || limit > 100 {
//		limit = 10
//	}
//	if err != nil {
//		return c.Status(500).JSON("Invalid per_page option")
//	}
//	result, err := r.s.GetAllChats()
//	if err != nil {
//		log.Println(err)
//		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error getting all chats"})
//	}
//
//	return c.Status(fiber.StatusOK).JSON(fiber.Map{"chats": result})
//}

func (r *socialRoutes) EnterToChat(c *fiber.Ctx) error {
	var chatIdStringDTO entity.ChatIdStringDTO
	if err := c.BodyParser(&chatIdStringDTO); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	chatIDUUID, err := uuid.Parse(chatIdStringDTO.ChatID)

	var enterToChatDTO entity.EnterToChatDTO
	enterToChatDTO.ChatID = chatIDUUID
	enterToChatDTO.UserID = utils.GetUserIDFromContext(c)
	result, err := r.s.EnterToChat(&enterToChatDTO)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"result": result})
}

func (r *socialRoutes) CreateChat(c *fiber.Ctx) error {
	var createDTO entity.CreateChatDTO
	if err := c.BodyParser(&createDTO); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	createDTO.UserID = utils.GetUserIDFromContext(c)
	createdChat, err := r.s.CreateChat(&createDTO)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(createdChat)
}

func (r *socialRoutes) WebSocketHandler(c *websocket.Conn) {
	clientObj := pkg.ClientObject{
		GROUP: c.Locals("GROUP").(string),
		USER:  c.Locals("USER").(string),
		Conn:  c,
	}
	defer func() {
		pkg.Unregister <- clientObj
		c.Close()
	}()
	// Register the client
	pkg.Register <- clientObj
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}
			return // Calls the deferred function, i.e. closes the connection on error
		}
		if messageType == websocket.TextMessage {
			// Broadcast the received message
			pkg.Broadcast <- pkg.BroadcastObject{
				MSG:  string(message),
				FROM: clientObj,
			}
		} else {
			log.Println("websocket message received of type", messageType)
		}
	}
}
