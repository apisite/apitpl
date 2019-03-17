package tpl2x_test

import (
	"bytes"
	"html/template"
	"log"
	"os"

	"github.com/apisite/tpl2x"
	"github.com/apisite/tpl2x/lookupfs"
)

// Page holds page attributes
type Page struct {
	Title string
}

// SetTitle - set page title
func (p *Page) SetTitle(name string) (string, error) {
	p.Title = name
	return "", nil
}

// Here we demonstrate loading a set of templates from a directory.
func ExampleRenderResponse() {

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
		{[]string{"includes"}, "inc", `inc1`},
		{[]string{"includes", "subdir1"}, "inc", `inc2 here`},
		{[]string{"layouts"}, "default", `<title>{{ or .Title "Default title" }}</title>=={{ content -}}=={{ template "inc" .}} `},
		{[]string{"layouts", "subdir2"}, "lay", `lay2 here`},
		{[]string{"pages"}, "page", `page1 here`},
		{[]string{"pages", "subdir3"}, "page", `{{ .SetTitle "Template title" }}
page2 here ({{ template "subdir1/inc" .}})`},
	})
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)

	cfg.Root = dir
	tfs, err := tpl2x.New(tpl2x.Config{}).LookupFS(lookupfs.New(cfg)).Parse()
	if err != nil {
		log.Fatal(err)
	}
	var b bytes.Buffer
	page := &Page{}
	err = tfs.Execute(&b, "subdir3/page", template.FuncMap{}, page)
	if err != nil {
		log.Fatal(err)
	}
	b.WriteTo(os.Stdout)

	// Output:
	// <title>Template title</title>==
	// page2 here (inc2 here)==inc1

}
