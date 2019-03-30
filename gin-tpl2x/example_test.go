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

	"github.com/apisite/tpl2x/gin-tpl2x/sample"
)

func Example() {

	// BufferPool size for rendered templates
	const bufferSize int = 64

	l := logrus.New()
	log := mapper.NewLogger(l)

	allFuncs := make(template.FuncMap, 0)
	setProtoFuncs(allFuncs)

	cfg := lookupfs.Config{
		Includes:   "inc",
		Layouts:    "layout",
		Pages:      "page",
		Ext:        ".tmpl",
		DefLayout:  "default",
		Index:      "index",
		HidePrefix: ".",
	}
	// Here we attach an embedded filesystem
	fs := lookupfs.New(cfg).FileSystem(sample.FS())
	// Parse all of templates
	tfs, err := tpl2x.New(bufferSize).Funcs(allFuncs).LookupFS(fs).Parse()
	if err != nil {
		log.Fatal(err)
	}
	gintpl := gintpl2x.New(log, tfs)
	gintpl.RequestHandler = func(ctx *gin.Context, funcs template.FuncMap) gintpl2x.MetaData {
		setRequestFuncs(funcs, ctx)
		page := sample.NewMeta(http.StatusOK, "text/html; charset=utf-8")
		return &page
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	gintpl.Route("", r)

	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	fmt.Println(resp.Code)
	fmt.Println(resp.Header().Get("Content-Type"))
	fmt.Println(resp.Body.String())

	// Output:
	// 200
	// text/html; charset=utf-8
	// <html>
	// <head>
	//   <title>index page</title>
	// </head>
	// <body>
	//   <h2>Test data</h2>
	// <h3>My TODO list</h3>
	// <ul>
	// <li>Task 1
	// <li>Task 2
	// <li>Task 3
	// </ul>
	// <footer>
	// <hr>
	// Host: <br />
	// URL: /<br />
	// </footer>
	// </body>
	// </html>

}

// setProtoFuncs appends function templates and not related to request functions to funcs
func setProtoFuncs(funcs template.FuncMap) {
	funcs["data"] = func() interface{} { return nil }
	funcs["request"] = func() interface{} { return nil }
	funcs["param"] = func(key string) string { return "" }
	funcs["HTML"] = func(s string) template.HTML {
		return template.HTML(s)
	}
}

// setRequestFuncs appends funcs which return real data inside request processing
func setRequestFuncs(funcs template.FuncMap, ctx *gin.Context) {
	funcs["data"] = func() interface{} { return sample.Data }
	funcs["request"] = func() interface{} { return ctx.Request }
	funcs["param"] = func(key string) string { return ctx.Param(key) }
}
