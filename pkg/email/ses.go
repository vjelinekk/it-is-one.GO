package email

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

var (
	client    *ses.Client
	fromEmail string
)

// Init loads AWS config and sets the sender address.
// If from is empty, SendMissedDoseAlert falls back to log.Printf (local dev).
func Init(from string) {
	fromEmail = from
	if from == "" {
		return
	}
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("[SES] Failed to load AWS config: %v", err)
		return
	}
	client = ses.NewFromConfig(cfg)
	log.Printf("[SES] Initialized with sender: %s", from)
}

// VerifyEmail triggers AWS SES to send a verification email to the given address.
// This is required in sandbox mode before SES can send emails to that address.
func VerifyEmail(to string) {
	if client == nil {
		log.Printf("[SES] Skipping verification for %s (SES not configured)", to)
		return
	}
	_, err := client.VerifyEmailIdentity(context.Background(), &ses.VerifyEmailIdentityInput{
		EmailAddress: aws.String(to),
	})
	if err != nil {
		log.Printf("[SES] Failed to send verification to %s: %v", to, err)
		return
	}
	log.Printf("[SES] Verification email sent to %s", to)
}

// SendMissedDoseAlert sends a missed dose notification email via AWS SES.
// Falls back to log.Printf if SES is not configured.
func SendMissedDoseAlert(to, patientEmail, scheduledTime string) error {
	subject := "Missed Medication Alert"
	body := fmt.Sprintf(
		"Hello,\n\nYour patient %s missed their medication dose scheduled at %s.\n\nPlease check in with them.\n",
		patientEmail, scheduledTime,
	)

	if client == nil || fromEmail == "" {
		log.Printf("[EMAIL] To: %s — Patient %s missed their medication dose scheduled at %s",
			to, patientEmail, scheduledTime)
		return nil
	}

	_, err := client.SendEmail(context.Background(), &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data:    aws.String(body),
					Charset: aws.String("UTF-8"),
				},
			},
		},
		Source: aws.String(fromEmail),
	})
	if err != nil {
		log.Printf("[SES] Failed to send email to %s: %v", to, err)
		return err
	}
	log.Printf("[SES] Email sent to %s", to)
	return nil
}
