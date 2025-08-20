package templates

import (
	"embed"
	htmlTemplate "html/template"
	"log"
	"miniflux-digest/internal/models"
	textTemplate "text/template"
)

type EmailTemplateData struct {
	models.OverviewTemplateData
	URL string
	Summary string
}

//go:embed *.gohtml *.gotxt
var embedFS embed.FS

var (
	ArchiveTemplate *htmlTemplate.Template
	OverviewTemplate *htmlTemplate.Template
	EmailTemplate   *textTemplate.Template
)

func init() {
	var err error
	archiveTemplateName := "grouped-entries.gohtml"
	overviewTemplateName := "overview.gohtml"
	emailTemplateName := "email.gotxt"

	ArchiveTemplate, err = htmlTemplate.New(archiveTemplateName).Funcs(TemplateFuncs).ParseFS(embedFS, archiveTemplateName)

	if err != nil {
		log.Fatalf("Error parsing archive template: %v", err)
	}

	OverviewTemplate, err = htmlTemplate.New(overviewTemplateName).Funcs(TemplateFuncs).ParseFS(embedFS, overviewTemplateName)

	if err != nil {
		log.Fatalf("Error parsing overview template: %v", err)
	}

	EmailTemplate, err = textTemplate.New(emailTemplateName).Funcs(TemplateFuncs).ParseFS(embedFS, emailTemplateName)

	if err != nil {
		log.Fatalf("Error parsing email template: %v", err)
	}
}
