package utils

import (
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"tourism-backend/internal/entity"
)

func GenerateQRCode(purchaseID uuid.UUID) *entity.PurchaseQRDTO {
	url := fmt.Sprintf("./v1/tours/provider/%s/check", purchaseID.String())

	qrCodeBytes, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		// Handle error (could log or return a fallback)
		return nil
	}

	// Convert to base64 string
	qrCodeBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(qrCodeBytes)

	return &entity.PurchaseQRDTO{
		QRCode: qrCodeBase64,
		URL:    url,
	}
}
