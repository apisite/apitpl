package tpl2x_test

import (
	"bytes"
	"html/template"
	"log"
	"os"

	"github.com/apisite/tpl2x"
	"github.com/apisite/tpl2x/lookupfs"

	"github.com/apisite/tpl2x/sample"
)

// Render template with layout
func Example() {

	// BufferPool size for rendered templates
	const bufferSize int = 64

	cfg := lookupfs.Config{
		Includes:  "inc_minimal",
		Layouts:   "layouts",
		Pages:     "pages",
		Ext:       ".html",
		DefLayout: "default",
	}

	tfs, err := tpl2x.New(bufferSize).
		LookupFS(lookupfs.New(cfg).
			FileSystem(sample.FS())).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
	var b bytes.Buffer
	page := sample.NewMeta(0, "")
	err = tfs.Execute(&b, "subdir3/page", template.FuncMap{}, page)
	if err != nil {
		log.Fatal(err)
	}
	b.WriteTo(os.Stdout)

	// Output:
	// <title>Template title</title>
	// ==
	// page2 here (inc2 here)==inc1

}
