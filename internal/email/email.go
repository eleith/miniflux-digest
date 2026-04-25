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

func (s *EmailServiceImpl) Send(smtpConfig config.ConfigSmtp, digestConfig config.ConfigDigest, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.OverviewTemplateData) error {
	message := mail.NewMsg()
	client, err := mail.NewClient(
		smtpConfig.Host,
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithPort(smtpConfig.Port),
		mail.WithUsername(smtpConfig.User),
		mail.WithPassword(smtpConfig.Password))

	if err != nil {
		return err
	}

	if err := message.From(digestConfig.Email.From); err != nil {
		return err
	}

	if err := message.To(digestConfig.Email.To); err != nil {
		return err
	}

	subject := fmt.Sprintf("%s - %s", digestConfig.Title, data.GeneratedDate.Format("January 2, 2006"))
	
	var overviewURL string
	if len(data.Entries) > 0 {
		overviewURL = fmt.Sprintf("%s/archive/%s/%s/index.html", digestConfig.Host, data.DigestSlug, data.GeneratedDate.Format("2006-01-02"))
	}

	message.Subject(subject)

	emailData := templates.EmailTemplateData{
		OverviewTemplateData: *data,
		URL:                  overviewURL,
		AttachmentHint:       *digestConfig.Email.AttachFiles,
	}

	var body bytes.Buffer
	if err := s.EmailTemplate.Execute(&body, emailData); err != nil {
		return err
	}
	message.SetBodyString(mail.TypeTextPlain, body.String())

	if *digestConfig.Email.AttachFiles {
		for _, f := range groupedEntryFiles {
			message.AttachFile(f.Name(), mail.WithFileContentType("text/html"))
		}
	}

	return client.DialAndSend(message)
}