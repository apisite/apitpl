package tpl2x_test

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/apisite/tpl2x"
	"github.com/apisite/tpl2x/lookupfs"

	"github.com/apisite/tpl2x/sample"
)

// Handle set of templates via http
func Example_http() {

	// BufferPool size for rendered templates
	const bufferSize int = 64

	cfg := lookupfs.Config{
		Includes:  "includes",
		Layouts:   "layouts",
		Pages:     "pages",
		Ext:       ".html",
		DefLayout: "default",
	}

	funcs := template.FuncMap{
		"request": func() http.Request {
			return http.Request{}
		},
		"content": func() template.HTML { return template.HTML("") },
	}

	tfs, err := tpl2x.New(bufferSize).
		Funcs(funcs).
		LookupFS(
			lookupfs.New(cfg).
				FileSystem(sample.FS())).
		Parse()
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()
	for _, uri := range tfs.PageNames(false) {
		router.HandleFunc("/"+uri, handleHTML(tfs, uri))
	}

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/subdir3/page", nil)
	if err != nil {
		log.Fatal(err)
	}
	router.ServeHTTP(resp, req)

	fmt.Println(resp.Code)
	fmt.Println(resp.Header().Get("Content-Type"))
	fmt.Println(resp.Body.String())

	// Output:
	// 200
	// text/html; charset=utf-8
	// <title>Template title</title>
	// ==
	// page2 here (inc2 here)==inc1 (URI: /subdir3/page)

}

// handleHTML returns page handler
func handleHTML(tfs *tpl2x.TemplateService, uri string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//		log.Debugf("Handling page (%s)", uri)

		page := sample.NewMeta(http.StatusOK, "text/html; charset=utf-8")
		funcs := template.FuncMap{
			"request": func() http.Request {
				return *r
			},
		}
		content := tfs.RenderContent(uri, funcs, page)
		if page.Status() == http.StatusMovedPermanently || page.Status() == http.StatusFound {
			http.Redirect(w, r, page.Title, page.Status())
			return
		}
		header := w.Header()
		header["Content-Type"] = []string{page.ContentType()}
		w.WriteHeader(page.Status())
		funcs["content"] = func() template.HTML { return template.HTML(content.Bytes()) }

		err := tfs.Render(w, funcs, page, content)
		if err != nil {
			log.Fatal(err)
		}

	}
}
