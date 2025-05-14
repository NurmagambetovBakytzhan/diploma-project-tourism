package v1

import (
	"encoding/json"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/webhook"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
	"tourism-backend/internal/entity"
	"tourism-backend/internal/usecase"
	"tourism-backend/pkg/logger"
	"tourism-backend/pkg/payment"
	"tourism-backend/utils"
)

type tourismRoutes struct {
	t usecase.TourismInterface
	l logger.Interface
	p *payment.PaymentProcessor
}

func newTourismRoutes(handler *gin.RouterGroup, t usecase.TourismInterface, l logger.Interface, csbn *casbin.Enforcer, payment *payment.PaymentProcessor) {
	r := &tourismRoutes{t, l, payment}
	h := handler.Group("/tours")
	{
		user := h.Group("/users")
		user.Use(utils.JWTAuthMiddleware())
		{
			user.GET("/me", r.GetMe)
			user.POST("/like", r.LikeTour)
			user.POST("/avatar", r.AddAvatar)
			user.GET("/avatar", r.GetMyAvatar)
			user.GET("/get-purchase-qr/:id", r.GetPurchaseQR)
		}

		usertracking := h.Group("/")
		usertracking.Use(utils.JWTAuthMiddleware())
		{
			usertracking.GET("/tour-events/:id", r.GetTourEventByID)
		}

		pay := h.Group("/payment")
		pay.Use(utils.JWTAuthMiddleware())
		{
			pay.POST("/", r.PayTourEvent)
			pay.POST("/create-payment-intent", r.CreatePaymentIntent)
			pay.POST("/confirm-payment", r.ConfirmPayment)
			//pay.POST("/stripe-webhook", r.HandleStripeWebHook)
		}

		protected := h.Group("/provider")
		protected.Use(utils.JWTAuthMiddleware(), utils.CasbinMiddleware(csbn))
		{
			protected.POST("/", r.CreateTour)
			protected.POST("/tour-event", r.CreateTourEvent)
			protected.POST("/tour-category", r.CreateTourCategory)
			protected.POST("/tour-location", r.CreateTourLocation)
			protected.GET("/tour-location/:id", r.GetTourLocationByID)
			protected.POST("/:id/", r.AddFilesToTourByTourID)
			protected.PATCH("/", r.ChangeTour)
			protected.POST("/:id/check", r.CheckPurchase)
		}

		h.GET("/v1/tours/uploads/:type/:filename", r.GetStaticFiles)
		h.GET("/", r.GetTours)
		h.GET("/:id", r.GetTourByID)
		h.GET("/categories", r.GetAllCategories)
		h.GET("/tour-events", r.GetFilteredTourEvents)
		h.GET("/tour-events/:id/weather", r.GetWeatherByTourEventID)
	}
}

func (r *tourismRoutes) HandleStripeWebhook(c *gin.Context) {
	const endpointSecret = "whsec_your_webhook_secret"
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Handle successful payment
		fmt.Printf("Payment succeeded for %d!\n", paymentIntent.Amount)

	case "payment_intent.payment_failed":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Handle failed payment
		fmt.Printf("Payment failed: %v\n", paymentIntent.LastPaymentError)

	default:
		fmt.Printf("Unhandled event type: %s\n", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// CreatePaymentIntent godoc
// @Summary Create Payment Intent
// @Description Creation of Payment Intent for Stripe
// @Tags Payment
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.CreatePaymentIntentRequest "Purchase information"
// @Failure 400 {object} map[string]string "Invalid json body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/tours/payment/create-payment-intent [post]
// @Security Bearer
func (r *tourismRoutes) CreatePaymentIntent(c *gin.Context) {
	var req entity.CreatePaymentIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(req.Currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity.CreatePaymentIntentResponse{
		ClientSecret: pi.ClientSecret,
	})
}

func (r *tourismRoutes) ConfirmPayment(c *gin.Context) {
	var req entity.ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pi, err := paymentintent.Get(req.PaymentIntentID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         pi.Status,
		"amount":         pi.Amount,
		"currency":       pi.Currency,
		"payment_method": pi.PaymentMethod.ID,
		"receipt_email":  pi.ReceiptEmail,
	})
}

// CheckPurchase godoc
// @Summary Check QR Info
// @Description Checks whether the authenticated provider is the owner of the tour event associated with the given purchase ID
// @Tags Provider
// @Param id path string true "Purchase ID (UUID)"
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.Purchase "Purchase information"
// @Failure 400 {object} map[string]string "Invalid purchase ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "You are not the owner of this tour event"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/tours/provider/{id}/check [post]
// @Security Bearer
func (r *tourismRoutes) CheckPurchase(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	purchaseIDStr := c.Param("id")
	purchaseID, err := uuid.Parse(purchaseIDStr)
	if err != nil {
		log.Println("CheckPurchase err:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing purchase ID"})
		return
	}

	result, err := r.t.CheckPurchase(userID, purchaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetPurchaseQR godoc
// @Summary Get QR code for a purchase
// @Description Returns a QR code for the specified purchase ID if the user has access
// @Tags Users
// @Param id path string true "Purchase ID (UUID)"
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "QR code data"
// @Failure 400 {object} map[string]string "Invalid purchase ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Error getting purchase QR"
// @Router /v1/tours/users/get-purchase-qr/{id}/ [get]
// @Security Bearer
func (r *tourismRoutes) GetPurchaseQR(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	purchaseIDStr := c.Param("id")
	purchaseID, err := uuid.Parse(purchaseIDStr)
	if err != nil {
		log.Println("GetPurchaseQR err:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing purchase ID"})
		return
	}
	result, err := r.t.GetPurchaseQR(userID, purchaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting purchase QR"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetMyAvatar returns the avatar image info of the authenticated user.
//
// @Summary Get my avatar
// @Description Returns avatar metadata or path for the authenticated user.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Avatar info retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/tours/users/avatar [get]
// @Security Bearer
func (r *tourismRoutes) GetMyAvatar(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	result, err := r.t.GetMyAvatar(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user avatar"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
	return
}

// AddAvatar uploads avatar image for a specific tour.
//
// @Summary Upload avatar for a user
// @Description Allows authenticated users to upload avatar images.
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar Image"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Avatar uploaded successfully"
// @Failure 400 {object} map[string]string "Invalid request format or missing required fields"
// @Failure 403 {object} map[string]string "Unauthorized: You are not the owner of this tour"
// @Failure 500 {object} map[string]string "Failed to save file or database error"
// @Router /v1/tours/users/avatar/ [post]
// @Security Bearer
func (r *tourismRoutes) AddAvatar(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 200<<20) // 200MB

	if err := c.Request.ParseMultipartForm(200 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size too large"})
		return
	}
	avatar, exists := c.FormFile("avatar")
	if exists != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error uploading file"})
		return
	}
	filename := userID.String() + filepath.Ext(avatar.Filename)
	filespath := "./uploads/avatars/" + filename
	filespathToSave := "./v1/tours/uploads/avatars/" + filename

	if err := c.SaveUploadedFile(avatar, filespath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	err := r.t.SaveMyAvatar(userID, filespathToSave)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Avatar uploaded successfully", "result": filespathToSave})

}

// LikeTour godoc
// @Summary      Like a tour
// @Description  Allows a user to like a specific tour
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        likeTourDTO  body      entity.LikeTourDTO  true  "Tour ID to like"
// @Success      200          {object}  entity.UserFavorites
// @Failure 500 {object} map[string]string
// @Security     BearerAuth
// @Router       /v1/tours/users/like [post]
func (r *tourismRoutes) LikeTour(c *gin.Context) {
	var likeTourDTO entity.LikeTourDTO

	if err := c.ShouldBindJSON(&likeTourDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	tourIDUUID, err := uuid.Parse(likeTourDTO.TourID)
	result, err := r.t.LikeTour(userID, tourIDUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
	c.JSON(http.StatusOK, result)
}

// GetMe godoc
// @Summary Get current user info
// @Description Returns the information of the currently authenticated user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.User
// @Failure 500 {object} map[string]string
// @Router /v1/tours/users/me [get]
// @Security Bearer
func (r *tourismRoutes) GetMe(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)

	result, err := r.t.GetMe(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ChangeTour godoc
// @Summary Change an existing tour
// @Description Updates the details of a tour. Only the tour owner (provider) can modify their tours.
// @Tags Provider
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ChangeTour body entity.Tour true "Tour object to update"
// @Success 200 {object} entity.Tour "Updated tour info"
// @Failure 400 {object} map[string]string "Bad request - invalid body"
// @Failure 403 {object} map[string]string "Forbidden - not the owner"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/tours/provider [patch]
// @Security Bearer
func (r *tourismRoutes) ChangeTour(c *gin.Context) {
	var changeTour entity.Tour

	if err := c.ShouldBind(&changeTour); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !r.t.CheckTourOwner(changeTour.ID, utils.GetUserIDFromContext(c)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	result, err := r.t.ChangeTour(&changeTour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
	return
}

// AddFilesToTourByTourID uploads multiple panorama images for a specific tour.
//
// @Summary Upload panoramas for a tour
// @Description Allows authenticated tour providers to upload multiple panorama images for a specific tour.
// Only the owner of the tour can upload panoramas.
// @Tags Provider
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Tour ID (UUID)"
// @Param panoramas formData file true "Panorama images (can be multiple)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Images uploaded successfully"
// @Failure 400 {object} map[string]string "Invalid request format or missing required fields"
// @Failure 403 {object} map[string]string "Unauthorized: You are not the owner of this tour"
// @Failure 500 {object} map[string]string "Failed to save file or database error"
// @Router /v1/tours/provider/{id}/ [post]
// @Security Bearer
func (r *tourismRoutes) AddFilesToTourByTourID(c *gin.Context) {
	tourID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := utils.GetUserIDFromContext(c)

	if !r.t.CheckTourOwner(tourID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: You are not owner of this tour"})
		return
	}

	// Parse multiple files
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	files := form.File["panoramas"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No panoramas provided"})
		return
	}

	var panoramas []*entity.Panorama

	for _, file := range files {
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		filespath := "./uploads/panoramas/" + filename

		// Save the image file
		if err := c.SaveUploadedFile(file, filespath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		panoramas = append(panoramas, &entity.Panorama{
			PanoramaURL: filespath,
			TourID:      tourID,
		})
	}

	// Save to database
	result, err := r.t.AddFilesToTourByTourID(panoramas)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save panoramas to DB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Images uploaded successfully", "result": result})
}

// GetTourEventByID retrieves details of a specific tour event.
// @Summary Get tour event by ID
// @Description Fetches detailed information of a tour event by its ID
// @Tags Tour Events
// @Produce json
// @Param id path string true "Tour Event ID"
// @Success 200 {object} entity.TourEvent
// @Failure 500 {object} map[string]string
// @Router /v1/tours/tour-events/{id} [get]
func (r *tourismRoutes) GetTourEventByID(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)

	tourEventID, err := uuid.Parse(c.Param("id"))

	tourEvent, err := r.t.GetTourEventByID(tourEventID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	go r.t.TrackUserAction(userID, tourEvent.TourID)

	c.JSON(http.StatusOK, tourEvent)
}

// GetWeatherByTourEventID retrieves weather information for a specific tour event.
// @Summary Get weather by tour event ID
// @Description Fetches weather information related to a tour event
// @Tags Weather
// @Produce json
// @Param id path string true "Tour Event ID"
// @Success 200 {object} entity.WeatherInfo
// @Failure 500 {object} map[string]string
// @Router /v1/tours/tour-events/{id}/weather [get]
func (r *tourismRoutes) GetWeatherByTourEventID(c *gin.Context) {
	tourEventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	weatherInfo, err := r.t.GetWeatherByTourEventID(tourEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, weatherInfo)
}

// GetFilteredTourEvents retrieves filtered tour events based on query parameters.
// @Summary Get filtered tour events
// @Description Fetches a list of tour events filtered by criteria
// @Tags Tour Events
// @Produce json
// @Param start_date query string false "Start Date (YYYY-MM-DD)"
// @Param end_date query string false "End Date (YYYY-MM-DD)"
// @Param min_price query number false "Minimum Price"
// @Param max_price query number false "Maximum Price"
// @Success 200 {array} entity.TourEvent
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/tours/tour-events [get]
func (r *tourismRoutes) GetFilteredTourEvents(c *gin.Context) {
	var filter entity.TourEventFilter

	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	categoryIDs := c.QueryArray("category_ids")

	for _, id := range categoryIDs {
		parsedID, err := uuid.Parse(id)
		if err == nil {
			filter.CategoryIDs = append(filter.CategoryIDs, parsedID)
		}
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	if startDateStr != "" {
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		if err1 == nil {
			filter.StartDate = startDate
		}
	}
	if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			filter.EndDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}
	if minPrice := c.Query("min_price"); minPrice != "" {
		filter.MinPrice = utils.ParseFloat(minPrice)
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		filter.MaxPrice = utils.ParseFloat(maxPrice)
	}

	tourEvents, err := r.t.GetFilteredTourEvents(&filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tour events"})
		return
	}

	c.JSON(http.StatusOK, tourEvents)
}

// GetTourLocationByID retrieves a tour location by ID.
// @Summary Get tour location by ID
// @Description Fetches details of a specific tour location.
// @Tags Provider
// @Produce json
// @Param id path string true "Tour Location ID"
// @Security BearerAuth
// @Success 200 {object} entity.TourLocation "Tour location details"
// @Router /v1/tours/provider/tour-location/{id} [get]
// @Security Bearer
func (r *tourismRoutes) GetTourLocationByID(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)

	tourID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if !r.t.CheckTourOwner(tourID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: You are not owner of this tour"})
		return
	}
	tourLocation, err := r.t.GetTourLocationByID(tourID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tourLocation)

}

// CreateTourLocation creates a new tour location.
// @Summary Create a new tour location
// @Description Adds a new location for tours.
// @Tags Provider
// @Accept json
// @Produce json
// @Param location body entity.CreateTourLocationDTO true "Tour location details"
// @Security BearerAuth
// @Success 201 {object} entity.TourLocation "Created tour location"
// @Router /v1/tours/provider/tour-location [post]
// @Security Bearer
func (r *tourismRoutes) CreateTourLocation(c *gin.Context) {
	var createTourLocationDTO entity.CreateTourLocationDTO
	if err := c.ShouldBind(&createTourLocationDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if !r.t.CheckTourOwner(createTourLocationDTO.TourID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: You are not owner of this tour"})
		return
	}

	createdTourLocation, err := r.t.CreateTourLocation(&createTourLocationDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tour Location created successfully!", "Tour Location": createdTourLocation})
}

// CreateTourCategory creates a new tour category.
// @Summary Create a new tour category
// @Description Adds a new category for tours.
// @Tags Provider
// @Accept json
// @Produce json
// @Param category body entity.CreateTourCategoryDTO true "Tour category details"
// @Security BearerAuth
// @Success 201 {object} entity.TourCategory "Created tour category"
// @Router /v1/tours/provider/tour-category [post]
// @Security Bearer
func (r *tourismRoutes) CreateTourCategory(c *gin.Context) {
	var createTourCategoryDTO entity.CreateTourCategoryDTO
	if err := c.ShouldBind(&createTourCategoryDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if !r.t.CheckTourOwner(createTourCategoryDTO.TourID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: You are not owner of this tour"})
		return
	}

	createdTourCategory, err := r.t.CreateTourCategory(&createTourCategoryDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tour Category created successfully!", "Tour Category": createdTourCategory})

}

// GetAllCategories retrieves all tour categories.
// @Summary Get all categories
// @Description Fetches a list of all available tour categories
// @Tags Categories
// @Produce json
// @Success 200 {array} entity.Category
// @Failure 500 {object} map[string]string
// @Router /v1/tours/categories [get]
func (r *tourismRoutes) GetAllCategories(c *gin.Context) {
	categories, err := r.t.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tours"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Categories": categories})
}

// PayTourEvent handles tour event payments.
// @Summary Pay for a tour event
// @Description Processes a payment for a selected tour event
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerToken
// @Param request body entity.TourPurchaseRequest true "Purchase Request"
// @Success 200 {object} entity.Purchase
// @Failure 400 {object} map[string]string
// @Router /v1/tours/payment [post]
// @Security Bearer
func (r *tourismRoutes) PayTourEvent(c *gin.Context) {
	var purchaseRaw entity.TourPurchaseRequest
	if err := c.ShouldBindJSON(&purchaseRaw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	UserID := utils.GetUserIDFromContext(c)

	purchase := entity.Purchase{
		TourEventID: purchaseRaw.TourEventID,
		UserID:      UserID,
		Status:      "Processing",
	}

	processingPurchase, err := r.t.CreatePurchase(&purchase)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r.p.PurchaseQueue <- processingPurchase

	c.JSON(http.StatusOK, gin.H{"Purchase": processingPurchase})
}

// CreateTourEvent handles the creation of a new tour event related to some specific tour with images and videos.
// @Summary Create a new tour event
// @Description Create a new tour event.
// @Tags Provider
// @Accept json
// @Produce json
// @Param event body entity.CreateTourEventDTO true "Tour Event Details"
// @Success 201 {object} entity.CreateTourEventDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/tours/provider/tour-event [post]
// @Security Bearer
func (r *tourismRoutes) CreateTourEvent(c *gin.Context) {
	var createTourEventDTO entity.CreateTourEventDTO

	if err := c.ShouldBindJSON(&createTourEventDTO); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get User ID from JWTMiddleware
	userID := utils.GetUserIDFromContext(c)

	if !r.t.CheckTourOwner(createTourEventDTO.TourID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: You are not owner of this tour"})
		return
	}

	tour := &entity.TourEvent{
		TourID:         createTourEventDTO.TourID,
		Date:           createTourEventDTO.Date,
		Price:          createTourEventDTO.Price,
		Place:          createTourEventDTO.Place,
		AmountOfPlaces: createTourEventDTO.AmountOfPlaces,
		InstaPostURL:   createTourEventDTO.InstaPostURL,
	}

	createdTourEvent, err := r.t.CreateTourEvent(tour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tour. Make sure the tour with such ID exists"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tour event created successfully!", "Tour Event": createdTourEvent})

}

// GetStaticFiles serves static files (images and videos) for a given tour.
// @Summary Get static files for a tour
// @Description Fetches images and videos for a specific tour by ID.Example http://localhost:8080/uploads/videos/4f72a1cb-6ed4-4f01-b38b-b605d3062236.mp4.
// @Tags Tours
// @Produce json
// @Param id path string true "Tour ID"
// @Success 200 {object} map[string]interface{} "Returns a list of image and video URLs."
// @Failure 400 {object} map[string]string "Invalid Tour ID"
// @Failure 404 {object} map[string]string "Tour not found"
// @Router /v1/tours/{id} [get]
func (r *tourismRoutes) GetStaticFiles(c *gin.Context) {

}

// GetTourByID retrieves details of a specific tour.
// @Summary Get tour by ID
// @Description Fetches detailed information of a tour by its ID
// @Tags Tours
// @Produce json
// @Param id path string true "Tour ID"
// @Success 200 {object} entity.Tour
// @Failure 500 {object} map[string]string
// @Router /v1/tours/{id} [get]
func (r *tourismRoutes) GetTourByID(c *gin.Context) {
	tour, err := r.t.GetTourByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tour)
}

// GetTours retrieves a list of available tours.
// @Summary Get all tours
// @Description Fetches a list of all available tours
// @Tags Tours
// @Produce json
// @Success 200 {array} entity.Tour
// @Failure 500 {object} map[string]string
// @Router /v1/tours [get]
func (r *tourismRoutes) GetTours(c *gin.Context) {

	tours, err := r.t.GetTours()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tours"})
		return
	}

	c.JSON(http.StatusOK, tours)
}

// CreateTour handles the creation of a new tour with images and videos.
// @Summary Create a new tour
// @Description Create a new tour with images and videos.
// @Tags Provider
// @Accept multipart/form-data
// @Produce json
// @Param description formData string true "Tour Description"
// @Param route formData string true "Tour Route"
// @Param images formData file false "Tour Images (multiple allowed)"
// @Param videos formData file false "Tour Videos (multiple allowed)"
// @Success 201 {object} entity.TourDocs
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/tours/provider [post]
// @Security Bearer
func (r *tourismRoutes) CreateTour(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 200<<20) // 200MB

	if err := c.Request.ParseMultipartForm(200 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size too large"})
		return
	}

	// Get User ID from JWTMiddleware
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert user_id string to UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	description := c.PostForm("description")
	route := c.PostForm("route")
	name := c.PostForm("name")
	form, _ := c.MultipartForm()
	var imageFiles []*multipart.FileHeader
	var videoFiles []*multipart.FileHeader

	if files, exists := form.File["images"]; exists {
		imageFiles = files
	}
	if files, exists := form.File["videos"]; exists {
		videoFiles = files
	}

	tour := &entity.Tour{
		ID:          uuid.New(),
		Description: description,
		Route:       route,
		OwnerID:     userID,
		Name:        name,
	}

	createdTour, err := r.t.CreateTour(tour, imageFiles, videoFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tour"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tour created successfully", "tour": createdTour})

}
