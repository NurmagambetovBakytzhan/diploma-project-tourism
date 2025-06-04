package email

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/smtp"
	"os"

	"github.com/jordan-wright/email"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
}

func NewEmailService() *EmailService {
	return &EmailService{
		"smtp.gmail.com",                  // SMTP host
		"587",                             // SMTP port
		"nurmagambetovbakytzan@gmail.com", // SMTP username
		os.Getenv("google_password"),      // SMTP password
		"nurmagambetovbakytzan@gmail.com",
	}
}

func (es *EmailService) SendVerificationCode(to, code string) error {
	e := email.NewEmail()
	e.From = es.fromEmail
	e.To = []string{to}
	e.Subject = "Email Verification Code"
	e.HTML = []byte(fmt.Sprintf(`
		<html>
		<body>
			<h2>Your Verification Code</h2>
			<p>Please use the following code to verify your email address:</p>
			<h3>%s</h3>
			<p>This code will expire in 15 minutes.</p>
		</body>
		</html>
	`, code))

	auth := smtp.PlainAuth("", es.smtpUsername, es.smtpPassword, es.smtpHost)
	return e.Send(fmt.Sprintf("%s:%s", es.smtpHost, es.smtpPort), auth)
}

func GenerateVerificationCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, 4)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}
	return string(code), nil
}
