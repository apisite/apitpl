// Package ginapitpl implements a gin frontend for apitpl.
package ginapitpl

import (
	"bytes"
	"html/template"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/birkirb/loggers.v1"

	"github.com/apisite/apitpl"
)

// EngineKey holds gin context key name for engine storage
const EngineKey = "github.com/apisite/apitpl"

// MetaData holds template metadata access methods
type MetaData interface {
	apitpl.MetaData
	ContentType() string // Returns content type
	Location() string    // Returns redirect url
	Status() int         // Response status
}

// TemplateService allows to replace apitpl functionality with the other package
type TemplateService interface {
	PageNames(hide bool) []string
	Render(w io.Writer, funcs template.FuncMap, data apitpl.MetaData, content *bytes.Buffer) (err error)
	RenderContent(name string, funcs template.FuncMap, data apitpl.MetaData) *bytes.Buffer
}

// Template holds template engine attributes
type Template struct {
	RequestHandler func(ctx *gin.Context, funcs template.FuncMap) MetaData
	fs             TemplateService
	log            loggers.Contextual
}

// New creates template object
func New(log loggers.Contextual, fs TemplateService) *Template {
	return &Template{fs: fs, log: log}
}

// Middleware stores Engine in gin context
func (tmpl *Template) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(EngineKey, tmpl)
	}
}

// Route registers template routes into gin
func (tmpl Template) Route(prefix string, r *gin.Engine) {
	if prefix != "" {
		prefix = prefix + "/"
	}

	// we need this before page registering
	r.Use(tmpl.Middleware())

	for _, p := range tmpl.fs.PageNames(true) {
		r.GET(prefix+p, tmpl.handleHTML(p)) // TODO: map[content-type]Pages
	}
}

// handleHTML returns gin page handler
func (tmpl Template) handleHTML(uri string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if val, ok := ctx.Get(EngineKey); ok {
			if t, ok := val.(*Template); ok {
				t.HTML(ctx, uri)
				return
			}
		}
		tmpl.log.Error("Context without valid engine key", EngineKey)
	}
}

// HTML renders page for given uri with context
func (tmpl Template) HTML(ctx *gin.Context, uri string) {
	funcs := make(template.FuncMap)
	page := (tmpl.RequestHandler)(ctx, funcs)
	content := tmpl.fs.RenderContent(uri, funcs, page)
	if page.Status() == http.StatusMovedPermanently || page.Status() == http.StatusFound {
		ctx.Redirect(page.Status(), page.Location())
		return
	}
	r := renderer{fs: tmpl.fs, funcMap: funcs, data: page, content: content}
	ctx.Header("Content-Type", page.ContentType())
	ctx.Render(page.Status(), r)
}

// renderer holds per request rendering attributes
type renderer struct {
	fs      TemplateService
	content *bytes.Buffer
	funcMap template.FuncMap
	data    MetaData
}

// Render - render page and write it to w
func (r renderer) Render(w http.ResponseWriter) error {
	return r.fs.Render(w, r.funcMap, r.data, r.content)
}

// WriteContentType called when Status does not allow body
func (r renderer) WriteContentType(w http.ResponseWriter) {}
