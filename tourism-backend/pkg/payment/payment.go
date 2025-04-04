package payment

import (
	"log"
	"sync"
	"time"
	"tourism-backend/internal/entity"
	"tourism-backend/internal/usecase"
)

type PaymentProcessor struct {
	mu             sync.Mutex
	PurchaseQueue  chan *entity.Purchase
	tourismUsecase usecase.TourismInterface
}

func NewPaymentProcessor(bufferSize int, usecase usecase.TourismInterface) *PaymentProcessor {
	p := &PaymentProcessor{
		PurchaseQueue:  make(chan *entity.Purchase, bufferSize),
		tourismUsecase: usecase,
	}

	// Start the worker goroutine
	go p.ProcessPurchases()

	return p
}

func (p *PaymentProcessor) ProcessPurchases() {
	for purchase := range p.PurchaseQueue {
		p.mu.Lock()

		log.Printf("Processing purchase: User %s -> TourEvent %s\n", purchase.UserID, purchase.TourEventID)

		// Simulate payment processing
		time.Sleep(5 * time.Second) // Simulate network delay

		// Mock payment API
		success := mockPaymentGateway(1)

		if success {
			err := p.tourismUsecase.PayTourEvent(purchase)
			if err != nil {
				log.Printf("Payment processing error: %v\n", err)
				continue
			}
			log.Printf("Payment successful for User %s on TourEvent %s\n", purchase.UserID, purchase.TourEventID)
		} else {
			log.Printf("Payment failed for User %s\n", purchase.UserID)
		}

		p.mu.Unlock()
	}
}

// Simulation of a payment gateway response
func mockPaymentGateway(amount float64) bool {
	return amount > 0 // Simulate always successful payment
}
