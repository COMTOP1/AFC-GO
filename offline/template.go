package main

import (
	"crypto/rand"
	"embed"
	"fmt"
	"html/template"
	"io"
	"math/big"

	"github.com/microcosm-cc/bluemonday"
)

// tmpls are the storage of templates in the executable
//
//go:embed *.tmpl
var tmpls embed.FS

type Templater struct {
}

type Template string

const (
	OfflineTemplate Template = "offline.tmpl"
)

type TemplateType int

const (
	NoNavType TemplateType = iota
	RegularType
)

// NewTemplate returns the template format to be used
func NewTemplate() *Templater {
	return &Templater{}
}

// String returns the string equivalent of Template
func (t Template) String() string {
	return string(t)
}

func (t *Templater) RenderTemplate(w io.Writer, data interface{}, mainTmpl Template, templateType TemplateType) error {
	var err error

	t1 := template.New("_base.tmpl")

	t1.Funcs(t.getFuncMaps())

	switch templateType {
	case NoNavType:
		t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_topNoNav.tmpl", "_footer.tmpl", mainTmpl.String())
	case RegularType:
		t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_top.tmpl", "_footer.tmpl", mainTmpl.String())
	default:
		return fmt.Errorf("unable to parse template, invalid type: %d", templateType)
	}

	if err != nil {
		return fmt.Errorf("failed to get templates for template(RenderTemplate): %w", err)
	}

	return t1.Execute(w, data)
}

func (t *Templater) GetEmailTemplate(emailTemplate Template) (*template.Template, error) {
	return template.New(emailTemplate.String()).ParseFS(tmpls, emailTemplate.String())
}

// getFuncMaps returns all the in built functions that templates can use
func (t *Templater) getFuncMaps() template.FuncMap {
	p := bluemonday.NewPolicy()
	p.AllowElements("a", "ul", "ol", "li", "h2", "b", "i", "u", "strike", "div", "br", "p",
		"blockquote", "pre", "hr")
	p.AllowAttrs("class").OnElements("h2")
	p.AllowAttrs("href", "style").OnElements("a")
	p.AllowURLSchemes("mailto", "http", "https")
	p.RequireNoFollowOnLinks(false)

	// Justification - via inline style
	p.AllowAttrs("style").OnElements("div", "p", "h2", "span")

	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"inc": func(a int) int {
			return a + 1
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) float64 {
			return float64(a) / float64(b)
		},
		"randomImgInt": func() int64 {
			nBig, err := rand.Int(rand.Reader, big.NewInt(999999))
			if err != nil {
				panic(err)
			}
			return nBig.Int64()
		},
		"htmlTemplate": func(content string) template.HTML {
			safe := p.Sanitize(content)
			//nolint:gosec
			return template.HTML(safe)
		},
	}
}

// This section is for go template linter
var (
	AllTemplates = [][]string{
		{"offline.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
	}

	_ = AllTemplates
)
