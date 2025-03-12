package v1

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"mime/multipart"
	"net/http"
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
		h.GET("/", r.GetTours)
		h.GET("/:id", r.GetTourByID)
		h.GET("/categories", r.GetAllCategories)
		h.GET("/tour-events", r.GetFilteredTourEvents)
		h.GET("/tour-events/:id", r.GetTourEventByID)
		h.GET("/tour-events/:id/weather", r.GetWeatherByTourEventID)
		pay := h.Group("/payment")
		pay.Use(utils.JWTAuthMiddleware())
		{
			pay.POST("/", r.PayTourEvent)
		}

		protected := h.Group("/provider")
		protected.Use(utils.JWTAuthMiddleware(), utils.CasbinMiddleware(csbn))
		{
			protected.POST("/", r.CreateTour)
			protected.POST("/tour-event", r.CreateTourEvent)
			protected.POST("/tour-category", r.CreateTourCategory)
			protected.POST("/tour-location", r.CreateTourLocation)
			protected.GET("/tour-location/:id", r.GetTourLocationByID)
		}
	}
}

func (r *tourismRoutes) GetTourEventByID(c *gin.Context) {
	tourEventID, err := uuid.Parse(c.Param("id"))

	tourEvent, err := r.t.GetTourEventByID(tourEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tourEvent)
}

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

func (r *tourismRoutes) GetFilteredTourEvents(c *gin.Context) {
	var filter entity.TourEventFilter

	if err := c.BindQuery(&filter); err != nil {
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
	if startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		endDate, err2 := time.Parse("2006-01-02", endDateStr)
		if err1 == nil && err2 == nil {
			filter.StartDate = startDate
			filter.EndDate = endDate
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

func (r *tourismRoutes) GetAllCategories(c *gin.Context) {
	categories, err := r.t.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tours"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Categories": categories})
}

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

func (r *tourismRoutes) CreateTourEvent(c *gin.Context) {
	var createTourEventDTO entity.CreateTourEventDTO

	if err := c.ShouldBindJSON(&createTourEventDTO); err != nil {
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
	}

	createdTourEvent, err := r.t.CreateTourEvent(tour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tour. Make sure the tour with such ID exists"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tour event created successfully!", "Tour Event": createdTourEvent})

}

func GetStaticFiles(c *gin.Context) {

}

func (r *tourismRoutes) GetTourByID(c *gin.Context) {
	tour, err := r.t.GetTourByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tour)
}

func (r *tourismRoutes) GetTours(c *gin.Context) {

	tours, err := r.t.GetTours()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tours"})
		return
	}

	c.JSON(http.StatusOK, tours)
}

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
	}

	createdTour, err := r.t.CreateTour(tour, imageFiles, videoFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tour"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tour created successfully", "tour": createdTour})

}
