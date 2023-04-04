package apitpl_test

import (
	"bytes"
	"embed"
	"io/fs"
	"html/template"
	"log"
	"os"

	"github.com/apisite/apitpl"
	"github.com/apisite/apitpl/lookupfs"

	"github.com/apisite/apitpl/samplemeta"
)

//go:embed testdata/*
var embedFS embed.FS

// Render template with layout
func Example_execute() {

	// BufferPool size for rendered templates
	const bufferSize int = 64

	cfg := lookupfs.Config{
		Includes:  "inc_minimal",
		Layouts:   "layouts",
		Pages:     "pages",
		Ext:       ".html",
		DefLayout: "default",
	}
	embedDirFS,_ := fs.Sub(embedFS, "testdata")
	tfs, err := apitpl.New(bufferSize).
		LookupFS(lookupfs.New(cfg).
			FileSystem(embedDirFS)).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
	var b bytes.Buffer
	page := &samplemeta.Meta{}
	page.SetLayout("default")
	err = tfs.Execute(&b, "subdir3/page", template.FuncMap{}, page)
	if err != nil {
		log.Fatal(err)
	}
	_, err = b.WriteTo(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// <title>Template title</title>
	// ==
	// page2 here (inc2 here)==inc1

}
