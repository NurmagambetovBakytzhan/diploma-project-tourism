package v1

import (
	"authentication-service/internal/entity"
	"authentication-service/internal/usecase"
	"authentication-service/pkg/logger"
	"authentication-service/utils"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"net/http"
)

type userRoutes struct {
	t usecase.UserInterface
	l logger.Interface
}

func newUserRoutes(handler *gin.RouterGroup, t usecase.UserInterface, l logger.Interface, csbn *casbin.Enforcer) {
	r := &userRoutes{t, l}

	h := handler.Group("/users")
	{
		//h.GET("/", r.GetTours)
		h.POST("/", r.RegisterUser)
		h.POST("/login", r.LoginUser)
		h.POST("/verify", r.VerifyEmail)

		protected := h.Group("/")
		protected.Use(utils.JWTAuthMiddleware())
		{
			protected.GET("/me", r.GetMe)
		}
	}
	adm := handler.Group("/admin")
	adm.Use(utils.JWTAuthMiddleware(), utils.CasbinMiddleware(csbn))
	{
		adm.GET("/users", r.GetUsers)
	}
}

// GetMe godoc
// @Summary Get current user info
// @Description Returns the information of the currently authenticated user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.User
// @Failure 500 {object} map[string]string
// @Router /v1/users/me [get]
// @Security Bearer
func (r *userRoutes) GetMe(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)

	result, err := r.t.GetMe(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, result)
}

func (r *userRoutes) GetUsers(c *gin.Context) {
	users, err := r.t.GetUsers()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, &users)
}

// @Summary User login
// @Description Authenticates a user and returns a JWT token.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body entity.LoginUserDTO true "Login credentials"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Authentication error"
// @Router /v1/users/login [post]
func (r *userRoutes) LoginUser(c *gin.Context) {
	var input entity.LoginUserDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := r.t.LoginUser(&input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// @Summary Register a new user
// @Description Creates a new user account with a hashed password.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body entity.CreateUserDTO true "User registration data"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Error hashing password"
// @Router /v1/users/ [post]
func (r *userRoutes) RegisterUser(c *gin.Context) {
	var createUserDTO entity.CreateUserDTO
	if err := c.ShouldBindJSON(&createUserDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := utils.HashPassword(createUserDTO.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	user := entity.User{
		Username: createUserDTO.Username,
		Email:    createUserDTO.Email,
		Password: hashedPassword,
		Role:     "user",
	}

	createdUser, sessionID, err := r.t.RegisterUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "User registered successfully. Please check your email for verification code.",
		"session_id": sessionID,
		"user":       createdUser,
	})
}

func (r *userRoutes) VerifyEmail(c *gin.Context) {
	var verifyDTO entity.VerifyEmailDTO
	if err := c.ShouldBindJSON(&verifyDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := r.t.VerifyEmail(verifyDTO.SessionID, verifyDTO.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}
