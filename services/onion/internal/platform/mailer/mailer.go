package mailer

import (
	"context"
	"encoding/pem"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/toorop/go-dkim"
	gomail "gopkg.in/gomail.v2"
)

type Mailer struct {
	host       string
	port       int
	username   string
	password   string
	privateKey []byte
	domain     string
	selector   string
}

func NewMailer() (*Mailer, error) {
	privateKeyStr := os.Getenv("DKIM_PRIVATE_KEY")
	if privateKeyStr == "" {
		return nil, ErrDKIMPrivateKeyNotSet
	}

	host := os.Getenv("NODEMAILER_HOST")
	if host == "" {
		return nil, ErrNodemailerHostNotSet
	}

	portStr := os.Getenv("NODEMAILER_PORT")
	if portStr == "" {
		return nil, ErrNodemailerPortNotSet
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid NODEMAILER_PORT: %w", err)
	}

	username := os.Getenv("NODEMAILER_USER")
	if username == "" {
		return nil, ErrNodemailerUserNotSet
	}

	password := os.Getenv("NODEMAILER_PASS")
	if password == "" {
		return nil, ErrNodemailerPassNotSet
	}

	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return nil, ErrFailedToParsePEM
	}

	return &Mailer{
		host:       host,
		port:       port,
		username:   username,
		password:   password,
		privateKey: block.Bytes,
		domain:     "vbitez.voxire.com",
		selector:   "default",
	}, nil
}

func (m *Mailer) SendEmail(ctx context.Context, to, subject, body string) error {
	from := fmt.Sprintf("no-reply <%s>", m.username)

	isHTML := strings.Contains(body, "<html") || strings.Contains(body, "<body") || strings.Contains(body, "<p")

	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)

	if isHTML {
		msg.SetBody("text/html", body)
	} else {
		msg.SetBody("text/plain", body)
	}

	var mimeBuilder strings.Builder

	_, err := msg.WriteTo(&mimeBuilder)
	if err != nil {
		return fmt.Errorf("failed to build MIME: %w", err)
	}

	rawMessage := []byte(mimeBuilder.String())

	options := dkim.NewSigOptions()
	options.PrivateKey = m.privateKey
	options.Domain = m.domain
	options.Selector = m.selector
	options.Headers = []string{"From", "To", "Subject"}
	options.AddSignatureTimestamp = true
	options.Canonicalization = "relaxed/relaxed"

	err = dkim.Sign(&rawMessage, options)
	if err != nil {
		return fmt.Errorf("failed to DKIM sign: %w", err)
	}

	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	const maxRetries = 3

	var lastErr error

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			lastErr = smtp.SendMail(addr, auth, m.username, []string{to}, rawMessage)
			if lastErr == nil {
				return nil
			}

			time.Sleep(1 * time.Second)
		}
	}

	return fmt.Errorf("failed to send email after retries: %w", lastErr)
}

var (
	ErrDKIMPrivateKeyNotSet = errors.New("DKIM_PRIVATE_KEY environment variable is not set")
	ErrNodemailerHostNotSet = errors.New("NODEMAILER_HOST environment variable is not set")
	ErrNodemailerPortNotSet = errors.New("NODEMAILER_PORT environment variable is not set")
	ErrNodemailerUserNotSet = errors.New("NODEMAILER_USER environment variable is not set")
	ErrNodemailerPassNotSet = errors.New("NODEMAILER_PASS environment variable is not set")
	ErrFailedToParsePEM     = errors.New("failed to parse PEM block for private key")
)
