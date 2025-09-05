package email

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/templates"

	"github.com/wneessen/go-mail"
)

type EmailServiceImpl struct{
	EmailTemplate *template.Template
}

func (s *EmailServiceImpl) Send(cfg *config.Config, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.OverviewTemplateData) error {
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

	subject := fmt.Sprintf("Your Miniflux Digest for %s", data.GeneratedDate.Format("January 2, 2006"))
	overviewURL := fmt.Sprintf("%s/archive/%s/index.html", cfg.Digest.Host, data.GeneratedDate.Format("2006-01-02"))

	message.Subject(subject)

	emailData := templates.EmailTemplateData{
		OverviewTemplateData: *data,
		URL:                  overviewURL,
		
	}

	var body bytes.Buffer
	if err := s.EmailTemplate.Execute(&body, emailData); err != nil {
		return err
	}
	message.SetBodyString(mail.TypeTextPlain, body.String())

	for _, f := range groupedEntryFiles {
		message.AttachFile(f.Name(), mail.WithFileContentType("text/html"))
	}

	return client.DialAndSend(message)
}
