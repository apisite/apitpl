package gintpl2x_test

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"

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
	ContentType string
	Status      int
}

// SetTitle - set page title
func (p *Meta) SetTitle(name string) (string, error) {
	p.Title = name
	return "", nil
}

func ExampleNew() {

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
		Root:      "./tmpl",
	}

	fs := lookupfs.New(cfg)
	tfs, err := tpl2x.New(tpl2x.Config{}).Funcs(allFuncs).LookupFS(fs).Parse()
	if err != nil {
		log.Fatal(err)
	}
	gintpl := gintpl2x.New(log, tfs)
	gintpl.RequestHandler = requestHandler

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	gintpl.Route("", r)

	req, _ := http.NewRequest("GET", "/index", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	fmt.Println(resp.Code)
	fmt.Println(resp.Header().Get("Content-Type"))
	fmt.Println(resp.Body.String())

	// Output:
	//200
	//
	//<html>
	//<head>
	//   <title>index page</title>
	//</head>
	//<body>
	//   <a href="/">Home</a><br />
	//<h2>Demo pages</h2>
	//<ul>
	//<li><a href="/admin/">admin</a>
	//<li><a href="/page">page</a>
	//<li><a href="/redir">redirect</a>
	//<li><a href="/page?err=on">error</a>
	//</ul>
	//gin-mulate addons:
	//<ul>
	//<li><a href="/my/gopher/hello">var from url</a>
	//</ul>
	//<h2>Test data</h2>
	//<h3>My TODO list</h3>
	//<ul>
	//<li>Task 1
	//<li>Task 2
	//<li>Task 3
	//</ul>
	//<footer>
	//<hr>
	//Host: <br />
	//URL: /index<br />
	//</footer>
	//</body>
	//</html>

}

// funcs which return real data inside request processing
func requestHandler(ctx *gin.Context, funcs template.FuncMap) interface{} {
	funcs["data"] = func() interface{} { return data }
	funcs["request"] = func() interface{} { return ctx.Request }
	funcs["param"] = func(key string) string { return ctx.Param(key) }

	page := Meta{Status: http.StatusOK, ContentType: "text/html; charset=utf-8"}
	return &page
}

// SetFuncBlank appends function templates and not related to request functions to funcs
func SetFuncBlank(funcs template.FuncMap) {
	funcs["data"] = func() interface{} { return nil }
	funcs["request"] = func() interface{} { return nil }
	funcs["param"] = func(key string) string { return "" }
}
