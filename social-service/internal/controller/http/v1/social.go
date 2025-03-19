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
		chats := h.Group("/chats")
		{
			chats.Post("/", r.CreateChat)
			chats.Post("/enter", r.EnterToChat)
			chats.Get("/", r.GetMyChats)
			chats.Get("/:id/messages", r.GetChatMessages)
		}
	}
	wshandler := handler.Group("/ws")
	wshandler.Use(utils.JWTAuthMiddleware(), utils.WebSocketMiddleware())
	{
		wshandler.Get("/", websocket.New(r.WebSocketHandler))
	}

}

func (r *socialRoutes) GetChatMessages(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	chatID := c.Params("id")
	if chatID == "" {
		log.Println("chatID is empty")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "chat id is empty"})
	}
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "chat id is invalid"})
	}

	if !r.s.CheckChatParticipant(chatUUID, userID) {
		log.Printf("User: %s is not part of the chat: %s", userID, chatUUID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "You are not a member of this chat"})
	}

	result, err := r.s.GetChatMessages(chatUUID)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error getting messages"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"messages": result})
}

func (r *socialRoutes) GetMyChats(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	result, err := r.s.GetMyChats(userID)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error getting all chats"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"chats": result})
}

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
		ChatID: c.Locals("ChatID").(string),
		UserID: c.Locals("UserID").(string),
		Conn:   c,
	}
	defer func() {
		pkg.Unregister <- clientObj
		c.Close()
	}()
	if !r.s.CheckChatParticipant(utils.StringToUUID(clientObj.ChatID), utils.StringToUUID(clientObj.UserID)) {
		return
	}
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
			msg := string(message)

			go func() {
				messageToPost := entity.Message{
					Text:   msg,
					UserID: utils.StringToUUID(clientObj.UserID),
					ChatID: utils.StringToUUID(clientObj.ChatID),
				}
				r.s.PostMessage(&messageToPost)
			}()

			pkg.Broadcast <- pkg.BroadcastObject{
				MSG:  msg,
				FROM: clientObj,
			}
		} else {
			log.Println("websocket message received of type", messageType)
		}
	}
}
