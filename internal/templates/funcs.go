package templates

import (
	"embed"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed static
var staticFS embed.FS

var TemplateFuncs = template.FuncMap{
	"include":    include,
	"textToHTML": textToHTML,
	"htmlEscape": htmlEscape,
}

func include(filename string) (any, error) {
	b, err := fs.ReadFile(staticFS, filename)
	if err != nil {
		return "", err
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".css":
		return template.CSS(b), nil
	case ".js":
		return template.JS(b), nil
	default:
		return template.HTML(b), nil
	}
}

func textToHTML(s string) template.HTML {
	return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
}

func htmlEscape(s string) template.HTML {
	return template.HTML(s)
}
