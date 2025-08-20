package email

import (
	"fmt"
	"io"
	"os"

	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"

	"github.com/wneessen/go-mail"
)

type EmailServiceImpl struct{}

func (s *EmailServiceImpl) Send(cfg *config.Config, overviewFile *os.File, groupedEntryFiles []*os.File, data *models.HTMLTemplateData) error {
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

	subject := fmt.Sprintf("[miniflux digest] %s", data.OverviewSummary)

	message.Subject(subject)

	// Set the overview digest as the body
	overviewHTML, err := io.ReadAll(overviewFile)
	if err != nil {
		return err
	}
	message.SetBodyString(mail.TypeTextHTML, string(overviewHTML))

	// Attach all grouped digest HTML files
	for _, f := range groupedEntryFiles {
		message.AttachFile(f.Name(), mail.WithFileContentType("text/html"))
	}

	return client.DialAndSend(message)
}
