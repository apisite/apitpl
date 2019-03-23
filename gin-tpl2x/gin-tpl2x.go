// Package gintpl2x - tpl2x bindings for gin
package gintpl2x

// https://stackoverflow.com/questions/42747183/how-to-render-templates-to-multiple-layouts-in-go

import (
	"github.com/gin-gonic/gin"

	"bytes"
	"html/template"
	"net/http"

	"gopkg.in/birkirb/loggers.v1"

	"github.com/apisite/tpl2x" // TODO: change to interface
)

// EngineKey holds gin context key name for engine storage
const EngineKey = "github.com/apisite/tpl2x"

// MetaData holds template metadata access methods
type MetaData interface {
	tpl2x.MetaData
	ContentType() string // Return layout name
	Location() string    // Return redirect url
	Status() int         // Response status
}

// Template holds template engine attributes
type Template struct {
	fs             *tpl2x.TemplateService
	RequestHandler func(ctx *gin.Context, funcs template.FuncMap) MetaData
	log            loggers.Contextual
}

// New creates template object
func New(log loggers.Contextual, fs *tpl2x.TemplateService) *Template {
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

	for _, p := range tmpl.fs.PageNames() {
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
	funcs := make(template.FuncMap, 0)
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

// Renderer holds per request rendering attributes
type renderer struct {
	fs      *tpl2x.TemplateService
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
