package sms

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var client *sns.Client

// Init loads AWS config for SNS using the EC2 instance role.
func Init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("[SNS] Failed to load AWS config: %v", err)
		return
	}
	client = sns.NewFromConfig(cfg)
	log.Printf("[SNS] Initialized")
}

// SendSandboxOTP triggers AWS to send a verification OTP to the phone number (sandbox mode only).
func SendSandboxOTP(phone string) error {
	if client == nil {
		log.Printf("[SNS] Skipping sandbox OTP for %s (SNS not configured)", phone)
		return nil
	}
	_, err := client.CreateSMSSandboxPhoneNumber(context.Background(), &sns.CreateSMSSandboxPhoneNumberInput{
		PhoneNumber: aws.String(phone),
	})
	if err != nil {
		log.Printf("[SNS] Failed to send sandbox OTP to %s: %v", phone, err)
		return err
	}
	log.Printf("[SNS] Sandbox OTP sent to %s", phone)
	return nil
}

// VerifySandboxOTP confirms the OTP code for a phone number in sandbox mode.
func VerifySandboxOTP(phone, otp string) error {
	if client == nil {
		return nil
	}
	_, err := client.VerifySMSSandboxPhoneNumber(context.Background(), &sns.VerifySMSSandboxPhoneNumberInput{
		PhoneNumber: aws.String(phone),
		OneTimePassword: aws.String(otp),
	})
	if err != nil {
		log.Printf("[SNS] Failed to verify OTP for %s: %v", phone, err)
		return err
	}
	log.Printf("[SNS] Phone %s verified", phone)
	return nil
}

// Send sends an SMS to the given E.164 phone number.
// Falls back to log.Printf if SNS is not configured.
func Send(phone, message string) error {
	if client == nil {
		log.Printf("[SMS] To: %s — %s", phone, message)
		return nil
	}

	_, err := client.Publish(context.Background(), &sns.PublishInput{
		PhoneNumber: aws.String(phone),
		Message:     aws.String(message),
	})
	if err != nil {
		log.Printf("[SNS] Failed to send SMS to %s: %v", phone, err)
		return err
	}
	log.Printf("[SNS] SMS sent to %s", phone)
	return nil
}
