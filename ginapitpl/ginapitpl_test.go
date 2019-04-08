package ginapitpl

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"github.com/apisite/apitpl"
	"github.com/apisite/apitpl/lookupfs"

	"github.com/apisite/apitpl/ginapitpl/samplemeta"
)

func TestRender(t *testing.T) {

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
		"/page": `200
text/html; charset=utf-8
<html>
<head>
  <title>Test page</title>
</head>
<body>
  <a href="/">Home</a><br />
<h2>Test page</h2>
<h3>Page content</h3>
<footer>
<hr>
Host: <br />
URL: /page<br />
</footer>
</body>
</html>
`,
		"/page?wide=on": `200
text/html; charset=utf-8
<html>
<head>
  <title>Test page</title>
</head>
<body>
    <h2>Test page</h2>
<h3>Page content</h3>
<footer>
<hr>
Host: <br />
URL: /page?wide=on<br />
</footer>
</body>
</html>
`,
		"/page?err=on": `501
text/html; charset=utf-8
<html>
<head>
  <title>We got an error, sorry</title>
</head>
<body>
  <a href="/">Home</a><br />
Escaped &lt;b&gt;error&lt;/b&gt; description
    <footer>
<hr>
Host: <br />
URL: /page?err=on<br />
</footer>
</body>
</html>
`,
		"/err": `403
text/html; charset=utf-8
<html>
<head>
  <title>Error 403: Sorry</title>
</head>
<body>
  <a href="/">Home</a><br />
Error description
    <footer>
<hr>
Host: <br />
URL: /err<br />
</footer>
</body>
</html>
`,
		"/redir": `302
text/html; charset=utf-8
<a href="/page">Found</a>.

`,
		"/admin/": `200
text/html; charset=utf-8
<html>
<head>
  <title>admin index</title>
</head>
<body>
  <a href="/">Home</a><br />
<h2>admin index page</h2>
<footer>
<hr>
Host: <br />
URL: /admin/<br />
</footer>
</body>
</html>
`,
		"/my/777/hello": `200
text/html; charset=utf-8
<html>
<head>
  <title>Default title</title>
</head>
<body>
  <a href="/">Home</a><br />
<h2>Hello, 777!</h2>
<footer>
<hr>
Host: <br />
URL: /my/777/hello<br />
</footer>
</body>
</html>
`,
	}
	r := mkRouter()

	for k, v := range want {

		req, _ := http.NewRequest("GET", k, nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)
		got := fmt.Sprintf("%d\n%s\n%s", resp.Code, resp.Header().Get("Content-Type"), resp.Body.String())
		assert.Equal(t, v, got, fmt.Sprintf("Request for %s", k))
		//fmt.Println(got)
	}

}

func mkRouter() *gin.Engine {

	// BufferPool size for rendered templates
	const bufferSize int = 64

	l, _ := test.NewNullLogger()
	log := mapper.NewLogger(l)

	allFuncs := make(template.FuncMap)
	allFuncs["HTML"] = func(s string) template.HTML {
		return template.HTML(s)
	}
	setProtoFuncs(allFuncs)

	cfg := lookupfs.Config{
		Includes:   "inc",
		Layouts:    "layout",
		Pages:      "page",
		Ext:        ".tmpl",
		DefLayout:  "default",
		Index:      "index",
		Root:       "./testdata",
		HidePrefix: ".",
	}
	fs := lookupfs.New(cfg)
	tfs, err := apitpl.New(bufferSize).Funcs(allFuncs).LookupFS(fs).Parse()
	if err != nil {
		log.Fatal(err)
	}
	gintpl := New(log, tfs)
	gintpl.RequestHandler = func(ctx *gin.Context, funcs template.FuncMap) MetaData {
		setRequestFuncs(funcs, ctx)
		page := samplemeta.NewMeta(http.StatusOK, "text/html; charset=utf-8")
		return page
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	gintpl.Route("", r)
	return r
}

// setProtoFuncs appends function templates and not related to request functions to funcs
func setProtoFuncs(funcs template.FuncMap) {
	funcs["data"] = func() interface{} { return nil }
	funcs["request"] = func() interface{} { return nil }
	funcs["param"] = func(key string) string { return "" }
}

// setRequestFuncs appends funcs which return real data inside request processing
func setRequestFuncs(funcs template.FuncMap, ctx *gin.Context) {
	funcs["data"] = func() interface{} { return samplemeta.Data }
	funcs["request"] = func() interface{} { return ctx.Request }
	funcs["param"] = func(key string) string { return ctx.Param(key) }
}
