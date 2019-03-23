package gintpl2x_test

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"

	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/apisite/tpl2x"
	"github.com/apisite/tpl2x/gin-tpl2x"
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

func Example_index() {

	r := mkRouter()

	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	fmt.Println(resp.Code)
	fmt.Println(resp.Header().Get("Content-Type"))
	fmt.Println(resp.Body.String())

	// Output:
	//200
	//text/html; charset=utf-8
	//<html>
	//<head>
	//   <title>index page</title>
	//</head>
	//<body>
	//   <h2>Test data</h2>
	//<h3>My TODO list</h3>
	//<ul>
	//<li>Task 1
	//<li>Task 2
	//<li>Task 3
	//</ul>
	//<footer>
	//<hr>
	//Host: <br />
	//URL: /<br />
	//</footer>
	//</body>
	//</html>

}

func Example_page() {

	r := mkRouter()

	req, _ := http.NewRequest("GET", "/page", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	fmt.Println(resp.Code)
	fmt.Println(resp.Header().Get("Content-Type"))
	fmt.Println(resp.Body.String())

	// Output:
	//200
	//text/html; charset=utf-8
	//<html>
	//<head>
	//   <title>Test page</title>
	//</head>
	//<body>
	//   <a href="/">Home</a><br />
	//Page content
	//<footer>
	//<hr>
	//Host: <br />
	//URL: /page<br />
	//</footer>
	//</body>
	//</html>

}

func mkRouter() *gin.Engine {
	l := logrus.New()
	log := mapper.NewLogger(l)

	allFuncs := make(template.FuncMap, 0)
	allFuncs["HTML"] = func(s string) template.HTML {
		return template.HTML(s)
	}
	SetFuncBlank(allFuncs)

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
	gintpl := gintpl2x.New(log, tfs)
	gintpl.RequestHandler = requestHandler

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	gintpl.Route("", r)
	return r
}

// funcs which return real data inside request processing
func requestHandler(ctx *gin.Context, funcs template.FuncMap) gintpl2x.MetaData {
	funcs["data"] = func() interface{} { return data }
	funcs["request"] = func() interface{} { return ctx.Request }
	funcs["param"] = func(key string) string { return ctx.Param(key) }

	page := Meta{status: http.StatusOK, contentType: "text/html; charset=utf-8"}
	return &page
}

// SetFuncBlank appends function templates and not related to request functions to funcs
func SetFuncBlank(funcs template.FuncMap) {
	funcs["data"] = func() interface{} { return nil }
	funcs["request"] = func() interface{} { return nil }
	funcs["param"] = func(key string) string { return "" }
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
