package email

import (
	"fmt"
	"os"
	"path/filepath"

	"miniflux-digest/config"
	"miniflux-digest/internal/category"
	"miniflux-digest/internal/templates"

	"github.com/wneessen/go-mail"
)

type TextTemplateData struct {
	category.CategoryData
	URL string
}

func Send(cfg *config.Config, file *os.File, data *category.CategoryData) error {
	message := mail.NewMsg()
	client, err := mail.NewClient(
		cfg.SmtpHost,
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithPort(cfg.SmtpPort),
		mail.WithUsername(cfg.SmtpUser),
		mail.WithPassword(cfg.SmtpPassword))

	if err != nil {
		return err
	}

	if err := message.From(cfg.DigestEmailFrom); err != nil {
		return err
	}

	if err := message.To(cfg.DigestEmailTo); err != nil {
		return err
	}

	subject := fmt.Sprintf("[miniflux digest] %s", data.Category.Title)
	filename := filepath.Base(file.Name())
	dir := filepath.Base(filepath.Dir(file.Name()))
	url := fmt.Sprintf("%s/%s/%s", cfg.DigestHost, dir, filename)
	textData := TextTemplateData{
		CategoryData: *data,
		URL:          url,
	}

	message.Subject(subject)
	message.AttachFile(file.Name(), mail.WithFileContentType("text/html"))

	err = message.SetBodyTextTemplate(templates.EmailTemplate, textData)

	if err != nil {
		return err
	}

	return client.DialAndSend(message)
}
