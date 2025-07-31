package email

import (
	"fmt"
	"os"
	"path/filepath"

	"miniflux-digest/internal/config"
	"miniflux-digest/internal/app"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/templates"

	"github.com/wneessen/go-mail"
)


type EmailServiceImpl struct{}

var _ app.EmailService = (*EmailServiceImpl)(nil)

func (s *EmailServiceImpl) Send(cfg *config.Config, file *os.File, data *models.HTMLTemplateData) error {
	message := mail.NewMsg()
	client, err := mail.NewClient(
		cfg.Smtp.Host,
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithPort(cfg.Smtp.Port),
		mail.WithUsername(cfg.Smtp.User),
		mail.WithPassword(cfg.Smtp.Password))

	if err != nil {
		return err
	}

	if err := message.From(cfg.Digest.Email.From); err != nil {
		return err
	}

	if err := message.To(cfg.Digest.Email.To); err != nil {
		return err
	}

	subject := fmt.Sprintf("[miniflux digest] %s", data.Category.Title)
	filename := filepath.Base(file.Name())
	dir := filepath.Base(filepath.Dir(file.Name()))
	url := fmt.Sprintf("%s/%s/%s/%s", cfg.Digest.Host, "archive", dir, filename)
	textData := templates.EmailTemplateData{
		HTMLTemplateData: *data,
		URL:          url,
		Summary:      data.Summary,
	}

	message.Subject(subject)
	message.AttachFile(file.Name(), mail.WithFileContentType("text/html"))

	err = message.SetBodyTextTemplate(templates.EmailTemplate, textData)

	if err != nil {
		return err
	}

	return client.DialAndSend(message)
}
