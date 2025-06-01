package repo

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
	"tourism-backend/internal/entity"
	"tourism-backend/pkg/postgres"
)

const _defaultEntityCap = 64

// TourismRepo -.
type TourismRepo struct {
	PG *postgres.Postgres
}

// New -.
func NewTourismRepo(pg *postgres.Postgres) *TourismRepo {
	return &TourismRepo{pg}
}

func (r *TourismRepo) GetTourEventsByTourID(tourID uuid.UUID) ([]*entity.TourEvent, error) {
	var res []*entity.TourEvent

	err := r.PG.Conn.Model(&entity.TourEvent{}).Where("tour_id = ?", tourID).Find(&res).Error
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("get tour events by tour id")
	}
	return res, nil
}

func (r *TourismRepo) CheckPurchase(userID, purchaseID uuid.UUID) (*entity.Purchase, error) {
	var purchase entity.Purchase

	err := r.PG.Conn.Preload("TourEvent.Tour").
		First(&purchase, "id = ?", purchaseID).Error
	if err != nil {
		return nil, err
	}

	if purchase.TourEvent.Tour.OwnerID == userID {
		return &purchase, nil
	}

	return nil, nil
}

func (r *TourismRepo) GetPurchaseQR(userID, purchaseID uuid.UUID) (*entity.Purchase, error) {
	var purchase entity.Purchase
	err := r.PG.Conn.Model(&purchase).
		Where("id = ? AND user_id = ? AND status = ?", purchaseID, userID, "Paid").
		First(&purchase).Error
	if err != nil {
		log.Println("GetPurchaseQR err:", err)
		return nil, fmt.Errorf("tour event have not been paid or not exists")
	}
	return &purchase, nil
}

func (r *TourismRepo) SaveMyAvatar(userID uuid.UUID, avatar string) error {
	err := r.PG.Conn.Model(&entity.User{}).Where("id = ?", userID).Update("avatar_url", avatar).Error
	if err != nil {
		log.Println(err)
		return fmt.Errorf("save user avatar failed")
	}
	return nil
}

func (r *TourismRepo) GetMyAvatar(userID uuid.UUID) (string, error) {
	var result string
	err := r.PG.Conn.Model(&entity.User{}).Select("avatar_url").Where("id = ?", userID).First(&result).Error
	if err != nil {
		log.Println("GetMyAvatar err: ", err)
		return "", fmt.Errorf("error getting avatar: %w", err)
	}
	return result, nil
}

func (r *TourismRepo) CreateUserAction(userID, tourEventID uuid.UUID) {
	userAction := entity.UserActivity{
		UserID: userID,
		TourID: tourEventID,
	}
	err := r.PG.Conn.Create(&userAction).Error
	if err != nil {
		log.Printf("Error creating userAction: %v\n", err)
	}
	return
}

func (r *TourismRepo) LikeTour(userID uuid.UUID, tourID uuid.UUID) (*entity.UserFavorites, error) {
	var count int64

	tourFavorite := entity.UserFavorites{
		UserID: userID,
		TourID: tourID,
	}
	r.PG.Conn.Model(&tourFavorite).Where("user_id = ?").Count(&count)
	err := r.PG.Conn.Create(&tourFavorite).Error
	if err != nil {
		log.Println("LikeTour err:", err)
		return nil, fmt.Errorf("Error Liking Tour \n")
	}

	return &tourFavorite, nil
}

func (r *TourismRepo) GetMe(id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.PG.Conn.
		Preload("CreatedTours").
		Preload("PurchasedTourEvents").
		Preload("FavoriteTours").
		Preload("PurchasedTourEvents.TourEvent").
		Preload("PurchasedTourEvents.TourEvent.Tour").
		Preload("PurchasedTourEvents.TourEvent.Tour.TourImages").
		First(&user, id).
		Error
	if err != nil {
		log.Println("TourismRepo.GetMe: ", err)
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (r *TourismRepo) AddFileToTourByTourID(panoramaEntity []*entity.Panorama) ([]*entity.Panorama, error) {
	err := r.PG.Conn.Create(&panoramaEntity).Error
	if err != nil {
		return nil, err
	}
	return panoramaEntity, nil
}

func (r *TourismRepo) ChangeTour(tour *entity.Tour) (*entity.Tour, error) {
	err := r.PG.Conn.Model(&tour).Updates(tour).Error
	if err != nil {
		log.Println("TourismRepo ChangeTour err: ", err)
		return nil, fmt.Errorf("TourismRepo ChangeTour Error")
	}
	return tour, nil
}

func (r *TourismRepo) GetWeatherInfoByTourEventID(tourEventID uuid.UUID) (*entity.WeatherInfoRQ, error) {
	var result struct {
		Date      time.Time
		Time      time.Time
		Longitude float64
		Latitude  float64
	}
	err := r.PG.Conn.Raw(`
		SELECT e.date, l.longitude, l.latitude
		FROM tourism.tour_events AS e
		INNER JOIN tourism.tour_locations AS l ON e.tour_id = l.tour_id
		WHERE e.id = ?`, tourEventID).Scan(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no weather info found for tourEventID: %v", tourEventID)
		}
		return nil, err
	}

	formattedDate := result.Date.Format("2006-01-02") // YYYY-MM-DD
	formattedTime := result.Date.Format("15")
	return &entity.WeatherInfoRQ{
		Date:      formattedDate,
		Time:      formattedTime,
		Longitude: result.Longitude,
		Latitude:  result.Latitude,
	}, nil
}

func (r *TourismRepo) GetTourEventByID(id uuid.UUID) (*entity.TourEvent, error) {
	var tourEvent entity.TourEvent
	err := r.PG.Conn.First(&tourEvent, id).Error
	if err != nil {
		return nil, err
	}

	return &tourEvent, nil
}

func (r *TourismRepo) GetFilteredTourEvents(filter *entity.TourEventFilter) ([]*entity.TourEvent, error) {
	var tourEvents []*entity.TourEvent

	query := r.PG.Conn.
		Joins("JOIN tourism.tours ON tourism.tours.id = tourism.tour_events.tour_id").
		Joins("LEFT JOIN tourism.tour_categories ON tourism.tour_categories.tour_id = tourism.tours.id").
		Where("tourism.tour_events.is_opened = ?", true)

	// Filter by categories
	if len(filter.CategoryIDs) > 0 {
		query = query.Where("tourism.tour_categories.category_id IN ?", filter.CategoryIDs)
	}

	if !filter.StartDate.IsZero() {
		query = query.Where("tourism.tour_events.date >= ?", filter.StartDate)
	}

	// Filter by start date
	if !filter.EndDate.IsZero() {
		query = query.Where("DATE(tourism.tour_events.date) <= ?", filter.EndDate.Format("2006-01-02"))
	}

	// Filter by end date
	if !filter.EndDate.IsZero() {
		query = query.Where("DATE(tourism.tour_events.date) <= ?", filter.EndDate.Format("2006-01-02"))
	}

	// Filter by budget
	if filter.MinPrice > 0 {
		query = query.Where("tourism.tour_events.price >= ?", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		query = query.Where("tourism.tour_events.price <= ?", filter.MaxPrice)
	}

	// Execute the query
	err := query.Preload("Tour").Preload("Tour.TourImages").Find(&tourEvents).Error
	return tourEvents, err
}

func (r *TourismRepo) GetTourLocationByID(tourLocationID uuid.UUID) (*entity.TourLocation, error) {
	var tourLocation entity.TourLocation
	err := r.PG.Conn.Where("tour_id = ?", tourLocationID).First(&tourLocation).Error
	if err != nil {
		return nil, fmt.Errorf("get tour location by id: %w", err)
	}
	return &tourLocation, nil
}

func (r *TourismRepo) CreateTourLocation(tourLocation *entity.CreateTourLocationDTO) (*entity.TourLocation, error) {
	tourLocationEntity := &entity.TourLocation{
		TourID:    tourLocation.TourID,
		Longitude: tourLocation.Longitude,
		Latitude:  tourLocation.Latitude,
	}

	err := r.PG.Conn.Create(&tourLocationEntity).Error
	if err != nil {
		return nil, fmt.Errorf("Tour Location not Found")
	}
	return tourLocationEntity, nil

}

func (r *TourismRepo) GetAllCategories() ([]entity.Category, error) {
	var categories []entity.Category
	err := r.PG.Conn.Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("GetAllCategories: %w", err)
	}
	return categories, nil
}

func (r *TourismRepo) CreateTourCategory(tourCategory *entity.CreateTourCategoryDTO) (*entity.TourCategory, error) {

	category := &entity.TourCategory{
		TourID:     tourCategory.TourID,
		CategoryID: tourCategory.CategoryID,
	}

	// Check if the record already exists
	existingCategory := &entity.TourCategory{}
	err := r.PG.Conn.Where("tour_id = ? AND category_id = ?", category.TourID, category.CategoryID).First(existingCategory).Error
	if err == nil {
		return existingCategory, fmt.Errorf("tour Category already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("Tour Category not Found")
	}

	err = r.PG.Conn.Create(&category).Error
	if err != nil {
		return nil, fmt.Errorf("Tour Category or Tour not Found")
	}
	return category, nil
}

func (r *TourismRepo) CreatePurchase(purchase *entity.Purchase) (*entity.Purchase, error) {
	err := r.PG.Conn.Transaction(func(tx *gorm.DB) error {
		// Check tour event record in the database
		var tourEvent entity.TourEvent

		// Fetch the tour event to verify conditions before updating
		if err := tx.Where("id = ? AND is_opened = ? AND amount_of_places > 0",
			purchase.TourEventID, true).First(&tourEvent).Error; err != nil {
			return fmt.Errorf("tour event not found or closed: %w", err)
		}

		// Decrease the available places count
		if err := tx.Model(&entity.TourEvent{}).
			Where("id = ?", purchase.TourEventID).
			UpdateColumn("amount_of_places", gorm.Expr("amount_of_places - 1")).Error; err != nil {
			return fmt.Errorf("failed to update amount_of_places: %w", err)
		}

		// Create the purchase record in the database
		if err := tx.Create(purchase).Error; err != nil {
			return fmt.Errorf("create purchase failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("create purchase transaction failed: %w", err)
	}

	// Reload purchase with related data
	err = r.PG.Conn.Preload("User").Preload("TourEvent.Tour").
		First(purchase, "id = ?", purchase.ID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to preload purchase data: %w", err)
	}

	return purchase, nil
}

func (r *TourismRepo) PayTourEvent(purchase *entity.Purchase) *entity.Purchase {

	result := r.PG.Conn.Model(&purchase).
		Where("id = ? AND status = ?", purchase.ID, "Processing").
		Update("status", "Paid")

	if result.Error != nil {
		log.Printf("Error Processing tour event PayTourEvent: %s, \n ERROR:  %w \n", result, result.Error)

		err := r.PG.Conn.Model(&purchase).
			Where("id = ? AND status = ?", purchase.ID, "Paid").
			Update("status", "Failed").Error
		if err != nil {
			log.Println("Error Failing Purchase: ", err)
		}
		err = r.PG.Conn.Model(&purchase).
			Where("id = ?", purchase.TourEventID).
			UpdateColumn("amount_of_places", gorm.Expr("amount_of_places + 1")).Error
		if err != nil {
			log.Println("Error Increasing Amount of Places: ", err)
		}
		return nil
	}

	return purchase
}

func (r *TourismRepo) CheckTourOwner(tourID uuid.UUID, userID uuid.UUID) bool {
	var tourOwnerID string
	err := r.PG.Conn.Table("tourism.tours").
		Select("owner_id").
		Where("id = ?", tourID).
		Scan(&tourOwnerID).Error

	if err != nil {
		fmt.Println("Error checking tour owner:", err)
		return false
	}

	// Convert string to UUID
	ownerUUID, err := uuid.Parse(tourOwnerID)
	if err != nil {
		fmt.Println("Error parsing owner UUID:", err)
		return false
	}

	return ownerUUID == userID
}

func (r *TourismRepo) CreateTourEvent(tourEvent *entity.TourEvent) (*entity.TourEvent, error) {
	err := r.PG.Conn.Transaction(func(tx *gorm.DB) error {
		// Create the tour record in the database
		var count int64
		if err := tx.Model(&entity.Tour{}).Where("id = ?", tourEvent.TourID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("tour with id %s does not exist", tourEvent.TourID)
		}

		// Create the tour event record in the database
		if err := tx.Create(&tourEvent).Error; err != nil {
			return err
		}

		return nil
	})
	err = r.PG.Conn.Preload("Tour").First(tourEvent, "id = ?", tourEvent.ID).Error

	if err != nil {
		return nil, err
	}

	return tourEvent, nil
}

func (r *TourismRepo) GetTourByID(tourID string) (*entity.Tour, error) {
	var tour entity.Tour

	err := r.PG.Conn.Preload("TourImages").Preload("TourVideos").Preload("TourPanoramas").First(&tour, "id = ?", tourID).Error
	if err != nil {
		return nil, err
	}

	return &tour, nil
}

func (r *TourismRepo) GetTours() ([]entity.Tour, error) {
	var tours []entity.Tour
	err := r.PG.Conn.Preload("TourImages").Preload("TourVideos").Preload("TourPanoramas").Preload("TourEvents").Find(&tours).Error
	if err != nil {
		return nil, err
	}
	return tours, nil
}

func (r *TourismRepo) CreateTour(tour *entity.Tour, imageFiles []*multipart.FileHeader, videoFiles []*multipart.FileHeader) (*entity.Tour, error) {
	err := r.PG.Conn.Transaction(func(tx *gorm.DB) error {
		// Create the tour record in the database
		if err := tx.Create(&tour).Error; err != nil {
			return err
		}

		// Save images inside the transaction
		for _, file := range imageFiles {
			filename := uuid.New().String() + filepath.Ext(file.Filename)
			filespath := "./uploads/images/" + filename
			filespathToDB := "./v1/tours/uploads/images/" + filename
			// Save the image file
			if err := r.saveFile(file, filespath); err != nil {
				return err
			}
			image := &entity.Image{ImageURL: filespathToDB, TourID: tour.ID}
			if err := tx.Create(&image).Error; err != nil {
				return err
			}
		}

		// Save videos inside the transaction
		for _, file := range videoFiles {
			filename := uuid.New().String() + filepath.Ext(file.Filename)
			filespath := "./uploads/videos/" + filename
			filespathToDB := "./v1/tours/uploads/videos/" + filename
			// Save the video file
			if err := r.saveFile(file, filespath); err != nil {
				return err
			}
			// Append the video record to the list
			video := &entity.Video{VideoURL: filespathToDB, TourID: tour.ID}
			if err := tx.Create(&video).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tour, nil
}

// Helper function to save the file to the disk
func (r *TourismRepo) saveFile(file *multipart.FileHeader, path string) error {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the contents of the source file to the destination file
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

//
//// GetHistory -.
//func (r *TranslationRepo) GetHistory(ctx context.Context) ([]entity.Translation, error) {
//	sql, _, err := r.Builder.
//		Select("source, destination, original, translation").
//		From("history").
//		ToSql()
//	if err != nil {
//		return nil, fmt.Errorf("TranslationRepo - GetHistory - r.Builder: %w", err)
//	}
//
//	rows, err := r.Pool.Query(ctx, sql)
//	if err != nil {
//		return nil, fmt.Errorf("TranslationRepo - GetHistory - r.Pool.Query: %w", err)
//	}
//	defer rows.Close()
//
//	entities := make([]entity.Translation, 0, _defaultEntityCap)
//
//	for rows.Next() {
//		e := entity.Translation{}
//
//		err = rows.Scan(&e.Source, &e.Destination, &e.Original, &e.Translation)
//		if err != nil {
//			return nil, fmt.Errorf("TranslationRepo - GetHistory - rows.Scan: %w", err)
//		}
//
//		entities = append(entities, e)
//	}
//
//	return entities, nil
//}
//
//// Store -.
//func (r *TranslationRepo) Store(ctx context.Context, t entity.Translation) error {
//	sql, args, err := r.Builder.
//		Insert("history").
//		Columns("source, destination, original, translation").
//		Values(t.Source, t.Destination, t.Original, t.Translation).
//		ToSql()
//	if err != nil {
//		return fmt.Errorf("TranslationRepo - Store - r.Builder: %w", err)
//	}
//
//	_, err = r.Pool.Exec(ctx, sql, args...)
//	if err != nil {
//		return fmt.Errorf("TranslationRepo - Store - r.Pool.Exec: %w", err)
//	}
//
//	return nil
//}
