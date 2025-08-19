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

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"include": include,
	}
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
