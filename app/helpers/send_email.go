package helpers

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/mail.v2"
)

func SendEmailUser(clientID int64, toEmail, username, plainPassword, verificationLink string) error {
	host := os.Getenv("SMTP_HOST")
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return fmt.Errorf("invalid SMTP_PORT: %v", err)
	}

	senderEmail := os.Getenv("SMTP_EMAIL")
	senderPassword := os.Getenv("SMTP_PASSWORD")

	m := mail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Email Verification")

	emailBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<body>
			<p>Welcome! Your account has been created successfully.</p>
			<p><strong>Client ID:</strong> %d</p>
			<p><strong>Login Username:</strong> %s</p>
			<p><strong>Password:</strong> %s</p>
			<p>Please click the link below to verify your email address:</p>
			<p><a href="%s">%s</a></p>
			<p>If you didn't request this account, please ignore this email.</p>
			<br>
			<p>Best regards,<br>Client Portal Team</p>
		</body>
		</html>`, clientID, username, plainPassword, verificationLink, verificationLink)

	m.SetBody("text/html", emailBody)

	dialer := mail.NewDialer(host, port, senderEmail, senderPassword)
	dialer.StartTLSPolicy = mail.MandatoryStartTLS

	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Println("Verification email sent to:", toEmail)
	return nil
}

func SendUpdateNotificationEmail(clientName, toEmail, oldEmail, newEmail, username string) error {
	host := os.Getenv("SMTP_HOST")
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return fmt.Errorf("invalid SMTP_PORT: %v", err)
	}

	senderEmail := os.Getenv("SMTP_EMAIL")
	senderPassword := os.Getenv("SMTP_PASSWORD")

	m := mail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Email Update Notification")

	emailBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<body>
			<p>Hello <strong>%s</strong>,</p>
			<p>We want to inform you that your registered email address for the %s Portal has been updated.</p>
			<p><strong>Previous Email:</strong> %s</p>
			<p><strong>New Email:</strong> %s</p>
			<p>If you made this change, no further action is needed.</p>
			<p>If you did NOT request this change, please contact the administrator immediately.</p>
			<br>
			<p>Best regards,<br>Yodu Team</p>
		</body>
		</html>`, username, clientName, oldEmail, newEmail)

	m.SetBody("text/html", emailBody)

	dialer := mail.NewDialer(host, port, senderEmail, senderPassword)
	dialer.StartTLSPolicy = mail.MandatoryStartTLS

	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send update notification email: %v", err)
	}

	log.Println("Email update notification sent to:", toEmail)
	return nil
}

func SendPasswordChangedEmail(toEmail, username string) error {
	host := os.Getenv("SMTP_HOST")
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return fmt.Errorf("invalid SMTP_PORT: %v", err)
	}

	senderEmail := os.Getenv("SMTP_EMAIL")
	senderPassword := os.Getenv("SMTP_PASSWORD")

	m := mail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Password Changed Notification")

	emailBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<body>
			<p>Hi %s,</p>
			<p>This is a confirmation that your password has been successfully changed.</p>
			<p>If you did not perform this change, please contact support immediately.</p>
			<br>
			<p>Best regards,<br>Yodu Team</p>
		</body>
		</html>`, username)

	m.SetBody("text/html", emailBody)

	dialer := mail.NewDialer(host, port, senderEmail, senderPassword)
	dialer.StartTLSPolicy = mail.MandatoryStartTLS

	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send password changed email: %v", err)
	}

	log.Println("Password change confirmation email sent to:", toEmail)
	return nil
}

func SendPasswordResetEmail(toEmail, username, verificationLink string) error {
	host := os.Getenv("SMTP_HOST")
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return fmt.Errorf("invalid SMTP_PORT: %v", err)
	}

	senderEmail := os.Getenv("SMTP_EMAIL")
	senderPassword := os.Getenv("SMTP_PASSWORD")

	m := mail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Password Reset Request")

	emailBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				.button {
					background-color: #007bff;
					border: none;
					color: white;
					padding: 10px 20px;
					text-align: center;
					text-decoration: none;
					display: inline-block;
					font-size: 16px;
					margin: 10px 0;
					cursor: pointer;
					border-radius: 5px;
				}
			</style>
		</head>
		<body>
			<p>Hey, <strong>%s</strong></p>
			<p>You are receiving this email because we received a reset request from your account.</p>
			<p>
				<a href="%s" class="button">Reset Password</a>
			</p>
			<p>This password reset link will expire in 60 minutes.</p>
			<p>If you did not request a password reset, no further action is required.</p>
			<br>
			<p>Best regards,<br>Yodu Team</p>
		</body>
		</html>`, username, verificationLink)

	m.SetBody("text/html", emailBody)

	dialer := mail.NewDialer(host, port, senderEmail, senderPassword)
	dialer.StartTLSPolicy = mail.MandatoryStartTLS

	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Println("Password reset email sent to:", toEmail)
	return nil
}
