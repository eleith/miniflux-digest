package templates

import (
	"embed"
	htmlTemplate "html/template"
	"log"
	textTemplate "text/template"
)

//go:embed *.gohtml *.gotxt
var embedFS embed.FS

var (
	ArchiveTemplate *htmlTemplate.Template
	EmailTemplate   *textTemplate.Template
)

func init() {
	var err error
	archiveTemplateName := "entries.gohtml"
	emailTemplateName := "email.gotxt"

	ArchiveTemplate, err = htmlTemplate.New(archiveTemplateName).Funcs(htmlTemplate.FuncMap{
		"htmlEscape": func(s string) htmlTemplate.HTML {
			return htmlTemplate.HTML(s)
		},
	}).ParseFS(embedFS, archiveTemplateName)

	if err != nil {
		log.Fatalf("Error parsing archive template: %v", err)
	}

	EmailTemplate, err = textTemplate.New(emailTemplateName).ParseFS(embedFS, emailTemplateName)

	if err != nil {
		log.Fatalf("Error parsing email template: %v", err)
	}
}
