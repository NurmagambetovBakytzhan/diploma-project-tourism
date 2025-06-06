package usecase

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"log"
	"mime/multipart"
	"tourism-backend/internal/entity"
	"tourism-backend/internal/usecase/repo"
	"tourism-backend/utils"
)

// TranslationUseCase -.
type TourismUseCase struct {
	repo     *repo.TourismRepo
	producer sarama.SyncProducer
	//telegram *client.Client
}

// // NewTourismUseCase -.
//
//	func NewTourismUseCase(r *repo.TourismRepo, p sarama.SyncProducer, t *client.Client) *TourismUseCase {
//		return &TourismUseCase{
//			repo:     r,
//			producer: p,
//			telegram: t,
//		}
//	}
//
// NewTourismUseCase -.
func NewTourismUseCase(r *repo.TourismRepo, p sarama.SyncProducer) *TourismUseCase {
	return &TourismUseCase{
		repo:     r,
		producer: p,
	}
}

func (r *TourismUseCase) GetTourEventsByTourID(tourID uuid.UUID) ([]*entity.TourEvent, error) {
	return r.repo.GetTourEventsByTourID(tourID)
}
func (r *TourismUseCase) CheckPurchase(userID, purchaseID uuid.UUID) (*entity.Purchase, error) {
	result, err := r.repo.CheckPurchase(userID, purchaseID)
	if err != nil {
		log.Println("CheckPurchase err:", err)
		return nil, fmt.Errorf("check purchase error: %w", err)
	}
	if result == nil {
		log.Println("CheckPurchase result is nil")
		return nil, fmt.Errorf("you are not the owner of this tour event")
	}
	return result, nil
}

func (r *TourismUseCase) GetPurchaseQR(userID, purchaseID uuid.UUID) (*entity.PurchaseQRDTO, error) {
	_, err := r.repo.GetPurchaseQR(userID, purchaseID)
	if err != nil {
		log.Println("GetPurchaseQR err:", err)
		return nil, err
	}
	result := utils.GenerateQRCode(purchaseID)
	return result, nil
}

func (r *TourismUseCase) SaveMyAvatar(userID uuid.UUID, avatar string) error {
	return r.repo.SaveMyAvatar(userID, avatar)
}

func (r *TourismUseCase) GetMyAvatar(userID uuid.UUID) (string, error) {
	return r.repo.GetMyAvatar(userID)
}

func (r *TourismUseCase) TrackUserAction(userID uuid.UUID, tourEventID uuid.UUID) {
	r.repo.CreateUserAction(userID, tourEventID)
}

func (r *TourismUseCase) LikeTour(userID uuid.UUID, tourID uuid.UUID) (*entity.UserFavorites, error) {
	return r.repo.LikeTour(userID, tourID)
}

func (r *TourismUseCase) GetMe(id uuid.UUID) (*entity.User, error) {
	return r.repo.GetMe(id)
}

func (r *TourismUseCase) ChangeTour(tour *entity.Tour) (*entity.Tour, error) {
	return r.repo.ChangeTour(tour)
}

func (r *TourismUseCase) AddFilesToTourByTourID(panoramaEntity []*entity.Panorama) ([]*entity.Panorama, error) {
	return r.repo.AddFileToTourByTourID(panoramaEntity)
}

func (r *TourismUseCase) GetTourEventByID(id uuid.UUID) (*entity.TourEvent, error) {
	return r.repo.GetTourEventByID(id)
}

func (r *TourismUseCase) GetWeatherByTourEventID(tourEventID uuid.UUID) (*entity.WeatherInfo, error) {
	tourWeatherRQ, err := r.repo.GetWeatherInfoByTourEventID(tourEventID)
	if err != nil {
		return nil, fmt.Errorf("get tour weather info by tour event id: %w", err)
	}
	result, err := utils.GetWeatherInfo(tourWeatherRQ)
	return result, nil
}

func (r *TourismUseCase) GetFilteredTourEvents(filter *entity.TourEventFilter) ([]*entity.TourEvent, error) {
	return r.repo.GetFilteredTourEvents(filter)
}

func (r *TourismUseCase) GetTourLocationByID(tourLocationID uuid.UUID) (*entity.TourLocation, error) {
	return r.repo.GetTourLocationByID(tourLocationID)
}

func (r *TourismUseCase) CreateTourLocation(tourLocation *entity.CreateTourLocationDTO) (*entity.TourLocation, error) {
	return r.repo.CreateTourLocation(tourLocation)
}

func (r *TourismUseCase) CreateTourCategory(tourCategory *entity.CreateTourCategoryDTO) (*entity.TourCategory, error) {
	return r.repo.CreateTourCategory(tourCategory)
}

func (r *TourismUseCase) GetAllCategories() ([]entity.Category, error) {
	categories, err := r.repo.GetAllCategories()
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (t *TourismUseCase) CreatePurchase(purchase *entity.Purchase) (*entity.Purchase, error) {
	return t.repo.CreatePurchase(purchase)
}

func (t *TourismUseCase) PayTourEvent(purchase *entity.Purchase) error {
	result := t.repo.PayTourEvent(purchase)
	if result == nil {
		log.Printf("Error paying tour event: ", result)
		return fmt.Errorf("Error paying tour event")
	}

	kafkaMessage := entity.Notification{
		Topic: "PAYMENT",
		Data: map[string]interface{}{
			"Text":    "Your payment processed successfully",
			"Payment": result,
		},
		Recipients: []uuid.UUID{purchase.UserID},
	}

	t.PublishMessage("notifications", kafkaMessage)

	return nil
}

func (t *TourismUseCase) CheckTourOwner(tourID uuid.UUID, userID uuid.UUID) bool {
	return t.repo.CheckTourOwner(tourID, userID)
}

func (t *TourismUseCase) CreateTourEvent(tourEvent *entity.TourEvent) (*entity.TourEvent, error) {
	//createNewBasicGroupChatRequest := client.CreateNewBasicGroupChatRequest{
	//	UserIds: []int64{597878414},
	//	Title:   tourEvent.Tour.Name,
	//}
	//chat, err := t.telegram.CreateNewBasicGroupChat(&createNewBasicGroupChatRequest)
	//if err != nil {
	//	log.Println("Usecase CreateGroupChat CreateTourEvent: ", err)
	//}
	//tourEvent.TelegramChatURL = strconv.FormatInt(chat.ChatId, 10)

	result, err := t.repo.CreateTourEvent(tourEvent)
	if err != nil {
		return nil, fmt.Errorf("create tour event: %w", err)
	}

	return result, nil
}

func (t *TourismUseCase) CreateTour(tour *entity.Tour, imageFiles []*multipart.FileHeader, videoFiles []*multipart.FileHeader) (*entity.Tour, error) {
	tour, err := t.repo.CreateTour(tour, imageFiles, videoFiles)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return tour, nil
}

func (t *TourismUseCase) GetTourByID(id string) (*entity.Tour, error) {
	tour, err := t.repo.GetTourByID(id)
	if err != nil {
		return nil, err
	}
	return tour, nil
}

func (t *TourismUseCase) GetTours() ([]entity.Tour, error) {
	tours, err := t.repo.GetTours()
	if err != nil {
		return nil, err
	}
	return tours, nil
}

func (t *TourismUseCase) PublishMessage(topic string, value interface{}) {
	jsonMessage, err := json.Marshal(value)
	if err != nil {
		log.Println("Failed to marshal message:", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(jsonMessage),
	}

	_, _, err = t.producer.SendMessage(msg)
	if err != nil {
		log.Println("Failed to send message:", err)
	}
	log.Println("Message sent!")

}
