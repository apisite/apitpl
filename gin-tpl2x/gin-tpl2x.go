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

// Template holds template engine attributes
type Template struct {
	fs             *tpl2x.TemplateService
	RequestHandler func(ctx *gin.Context, funcs template.FuncMap) interface{}
	log            loggers.Contextual
}

// New returns mulate template object
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
func (tmpl *Template) Route(prefix string, r *gin.Engine) {
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
func (tmpl *Template) handleHTML(uri string) gin.HandlerFunc {
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

// HTML renders page for given uri
func (tmpl *Template) HTML(ctx *gin.Context, uri string) {
	funcs := make(template.FuncMap, 0)
	// Get funcMap copy
	/*
		for k, v := range tmpl.fs.Funcs {
			funcs[k] = v
		}
	*/

	//if tmpl.FuncHandler != nil {
	page := (tmpl.RequestHandler)(ctx, funcs)
	//}
	//p, err := tmpl.RenderPage(uri, funcs, ctx.Request)

	content, err := tmpl.fs.RenderContent(uri, funcs, page)
	if err != nil {
		/*
			if p.Status == http.StatusMovedPermanently || p.Status == http.StatusFound {
				ctx.Redirect(p.Status, p.Title)
				return
			}
			tmpl.log.Debugf("page error: (%+v)", err)
			if p.Status == http.StatusOK {
				p.Status = http.StatusInternalServerError
				p.Raise(p.Status, "Internal", err.Error(), false)
			}
		*/
	}
	renderer := Renderer{fs: tmpl.fs, funcMap: funcs, data: page, content: content}
	//	ctx.Header("Content-Type", p.ContentType)

	//TODO	ctx.Render(p.Status, renderer)
	ctx.Render(http.StatusOK, renderer)
}

// Renderer holds per request rendering attributes
type Renderer struct {
	fs      *tpl2x.TemplateService
	content *bytes.Buffer
	funcMap template.FuncMap
	data    interface{}
}

// NewRenderer creates new renderer object
/*
func NewRenderer(fs *tpl2x.TemplateService, page *tpl2x.Page) *Renderer {
	return &Renderer{fs: fs, page: page}
}
*/
// Render - render page and write it to w
func (r Renderer) Render(w http.ResponseWriter) error {
	//	funcs["_content"] = func() template.HTML { return template.HTML(content.Bytes()) }
	return r.fs.Render(w, r.funcMap, r.data, r.content)
}

// WriteContentType writes page content type
func (r Renderer) WriteContentType(w http.ResponseWriter) {
	//header := w.Header()
	// TODO: r.Page.ContentType
	//	if val := header["Content-Type"]; len(val) == 0 {
	//TODO	header["Content-Type"] = []string{r.page.ContentType}
	//	}
}
