package digest

import (
	"bytes"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

func MinifyHTML(input []byte, enabled bool) ([]byte, error) {
	if !enabled {
		return input, nil
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFunc("text/html", html.Minify)

	var buf bytes.Buffer
	if err := m.Minify("text/html", &buf, bytes.NewReader(input)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
