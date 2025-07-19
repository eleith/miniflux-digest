package archive

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"miniflux-digest/internal/category"
	"miniflux-digest/internal/utils"
	"os"
)

var tmpl *template.Template

func init() {
	var err error
	templateName := "entries.gohtml"

	tmpl, err = template.New(templateName).Funcs(template.FuncMap{
		"htmlEscape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).ParseFiles("./templates/" + templateName)

	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}
}

func getHTML(data *category.CategoryData) (string, error) {
	var buf bytes.Buffer
	var htmlOutput string

	err := tmpl.Execute(&buf, data)

	if err == nil {
		htmlOutput = buf.String()
	}

	return htmlOutput, err
}

func makeArchiveFile(data *category.CategoryData) (*os.File, error) {
	categorySlug := utils.Slugify(data.Category.Title)
	categoryFolderPath := fmt.Sprintf("./web/miniflux-archive/%s", categorySlug)
	filename := fmt.Sprintf("%s/%s.html", categoryFolderPath, data.GeneratedDate.Format("2006-01-02"))
	err := os.MkdirAll(fmt.Sprintf("./%s", categoryFolderPath), os.ModePerm)

	if err == nil {
		file, err := os.Create(filename)
		return file, err
	}

	return nil, err
}

func MakeArchiveHTML(data *category.CategoryData) (*os.File, error) {
	file, err := makeArchiveFile(data)

	if err != nil {
		log.Printf("Error creating HTML file for category '%s': %v", data.Category.Title, err)
		return nil, err
	}

	htmlOutput, err := getHTML(data)

	if err != nil {
		log.Printf("Error generating HTML for category %s: %v", data.Category.Title, err)
		return file, err
	}

	_, err = file.WriteString(htmlOutput)

	if err != nil {
		log.Printf("Error writing HTML to file for category '%s': %v", data.Category.Title, err)
	}

	return file, err
}
