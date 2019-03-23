package tpl2x_test

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/apisite/tpl2x"
	"github.com/apisite/tpl2x/lookupfs"
)

// Handle set of templates via http
func Example_http() {

	cfg := lookupfs.Config{
		Includes:  "includes",
		Layouts:   "layouts",
		Pages:     "pages",
		Ext:       ".html",
		DefLayout: "default",
	}
	// Here we create a temporary directory and populate it with our sample
	// template definition files; usually the template files would already
	// exist in some location known to the program.
	dir := createTestDir(cfg.Ext, []templateFile{
		{[]string{"includes"}, "inc", `inc1 (URI: {{ request.URL }})`},
		{[]string{"includes", "subdir1"}, "inc", `inc2 here`},
		{[]string{"layouts"}, "default", `<title>{{ or .Title "Default title" }}</title>=={{ _content -}}=={{ template "inc" .}} `},
		{[]string{"layouts", "subdir2"}, "lay", `lay2 here`},
		{[]string{"pages"}, "page", `page1 here`},
		{[]string{"pages", "subdir3"}, "page", `{{ .SetTitle "Template title" }}
page2 here ({{ template "subdir1/inc" .}})`},
	})
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)

	cfg.Root = dir
	funcs := template.FuncMap{
		"request": func() http.Request {
			return http.Request{} //nil
		},
		"_content": func() template.HTML { return template.HTML("") },
	}
	tfs, err := tpl2x.New(bufferSize).Funcs(funcs).LookupFS(lookupfs.New(cfg)).Parse()
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()
	for _, uri := range tfs.PageNames() {
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
	// <title>Template title</title>==
	// page2 here (inc2 here)==inc1 (URI: /subdir3/page)

}

// handleHTML returns page handler
func handleHTML(tfs *tpl2x.TemplateService, uri string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//		log.Debugf("Handling page (%s)", uri)

		page := &Meta{Status: http.StatusOK, ContentType: "text/html; charset=utf-8"}
		funcs := template.FuncMap{
			"request": func() http.Request {
				return *r
			},
		}
		content := tfs.RenderContent(uri, funcs, page)
		if page.Status == http.StatusMovedPermanently || page.Status == http.StatusFound {
			http.Redirect(w, r, page.Title, page.Status)
			return
		}
		header := w.Header()
		header["Content-Type"] = []string{page.ContentType}
		w.WriteHeader(page.Status)
		funcs["_content"] = func() template.HTML { return template.HTML(content.Bytes()) }

		err := tfs.Render(w, funcs, page, content)
		if err != nil {
			log.Fatal(err)
		}

	}
}
