package gintpl2x

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/apisite/tpl2x"
	"github.com/apisite/tpl2x/lookupfs"
)

// Todo holds single todo item attrs
type Todo struct {
	Title string
	Done  bool
}

// TodoPageData holds todo page attrs
type TodoPageData struct {
	PageTitle string
	Todos     []Todo
}

var data = TodoPageData{
	PageTitle: "My TODO list",
	Todos: []Todo{
		{Title: "Task 1", Done: false},
		{Title: "Task 2", Done: true},
		{Title: "Task 3", Done: true},
	},
}

// Page holds page attributes
type Meta struct {
	Title       string
	contentType string
	status      int
	error       error
	layout      string
}

// SetTitle - set page title
func (p *Meta) SetTitle(name string) (string, error) {
	p.Title = name
	return "", nil
}
func (p *Meta) SetError(e error) {
	p.error = e
}

func (p Meta) Error() error        { return p.error }
func (p Meta) Layout() string      { return "default" }
func (p Meta) ContentType() string { return p.contentType }
func (p Meta) Status() int         { return p.status }
func (p Meta) Location() string    { return "" }

// templateFile defines the contents of a template to be stored in a file, for testing.
type templateFile struct {
	dirs     []string
	name     string
	contents string
}

// BufferPool size for rendered templates
const bufferSize int = 64

var templates = []templateFile{
	{[]string{"layout"}, "default", `<html>
<head>
  {{ template "header" . }}
</head>
<body>
  {{ template "menu" . -}}
  {{ content | HTML -}}
  {{ template "footer" . -}}
</body>
</html>
`},
	{[]string{"inc"}, "header", `<title>{{ or .Title "Default title" }}</title>`},
	{[]string{"inc"}, "footer", `<footer>
<hr>
Host: {{ request.Host }}<br />
URL: {{ request.URL.String | HTML }}<br />
</footer>
`},
	{[]string{"inc"}, "menu", `{{ if ne request.URL.String "/" -}}
<a href="/">Home</a><br />
{{ end -}}`},
	{[]string{"page"}, "index", `{{ .SetTitle "index page" -}}
<h2>Test data</h2>
<h3>{{ data.PageTitle }}</h3>
<ul>
{{range data.Todos -}}
    <li>{{- .Title }}
{{end -}}
</ul>
`},
	{[]string{"page"}, "page", `{{ .SetTitle "Test page" -}}Page content
`},
}

var want = map[string]string{
	"/": `200
text/html; charset=utf-8
<html>
<head>
  <title>index page</title>
</head>
<body>
  <h2>Test data</h2>
<h3>My TODO list</h3>
<ul>
<li>Task 1
<li>Task 2
<li>Task 3
</ul>
<footer>
<hr>
Host: <br />
URL: /<br />
</footer>
</body>
</html>
`,
	"page": `200
text/html; charset=utf-8
<html>
<head>
  <title>Test page</title>
</head>
<body>
  <a href="/">Home</a><br />
Page content
<footer>
<hr>
Host: <br />
URL: /page<br />
</footer>
</body>
</html>
`,
}

func TestRender(t *testing.T) {

	r := mkRouter()

	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)
	got := fmt.Sprintf("%d\n%s\n%s", resp.Code, resp.Header().Get("Content-Type"), resp.Body.String())
	assert.Equal(t, want["/"], got)

}

func mkRouter() *gin.Engine {
	l := logrus.New()
	log := mapper.NewLogger(l)

	allFuncs := make(template.FuncMap, 0)
	allFuncs["HTML"] = func(s string) template.HTML {
		return template.HTML(s)
	}
	setProtoFuncs(allFuncs)

	cfg := lookupfs.Config{
		Includes:  "inc",
		Layouts:   "layout",
		Pages:     "page",
		Ext:       ".tmpl",
		DefLayout: "default",
		Index:     "index",
		Root:      "./tmpl",
	}
	dir := createTestDir(cfg.Ext, templates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)
	cfg.Root = dir
	fs := lookupfs.New(cfg)
	tfs, err := tpl2x.New(bufferSize).Funcs(allFuncs).LookupFS(fs).Parse()
	if err != nil {
		log.Fatal(err)
	}
	gintpl := New(log, tfs)
	gintpl.RequestHandler = func(ctx *gin.Context, funcs template.FuncMap) MetaData {
		setRequestFuncs(funcs, ctx)
		page := Meta{status: http.StatusOK, contentType: "text/html; charset=utf-8"}
		return &page
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	gintpl.Route("", r)
	return r
}

func requestHandler(ctx *gin.Context, funcs template.FuncMap) MetaData {
	funcs["data"] = func() interface{} { return data }
	funcs["request"] = func() interface{} { return ctx.Request }
	funcs["param"] = func(key string) string { return ctx.Param(key) }

	page := Meta{status: http.StatusOK, contentType: "text/html; charset=utf-8"}
	return &page
}

// setProtoFuncs appends function templates and not related to request functions to funcs
func setProtoFuncs(funcs template.FuncMap) {
	funcs["data"] = func() interface{} { return nil }
	funcs["request"] = func() interface{} { return nil }
	funcs["param"] = func(key string) string { return "" }
}

// setRequestFuncs appends funcs which return real data inside request processing
func setRequestFuncs(funcs template.FuncMap, ctx *gin.Context) {
	funcs["data"] = func() interface{} { return data }
	funcs["request"] = func() interface{} { return ctx.Request }
	funcs["param"] = func(key string) string { return ctx.Param(key) }
}

func createTestDir(ext string, files []templateFile) string {
	dir, err := ioutil.TempDir("", "template")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		subpath := filepath.Join(file.dirs...)
		path := filepath.Join(dir, subpath)
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		f, err := os.Create(filepath.Join(path, file.name+ext))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		_, err = io.WriteString(f, file.contents)
		if err != nil {
			log.Fatal(err)
		}
	}
	return dir
}
