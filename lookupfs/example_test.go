package lookupfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	// "text/template"
)

// templateFile defines the contents of a template to be stored in a file, for testing.
type templateFile struct {
	dirs     []string
	name     string
	contents string
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

// Here we demonstrate loading a set of templates from a directory.
func ExampleLookupFilesByPrefix() {

	cfg := Config{
		Includes: "includes",
		Layouts:  "layouts",
		Pages:    "pages",
		Ext:      ".html",
	}
	// Here we create a temporary directory and populate it with our sample
	// template definition files; usually the template files would already
	// exist in some location known to the program.
	dir := createTestDir(cfg.Ext, []templateFile{
		{[]string{"includes"}, "inc", `inc1 here`},
		{[]string{"includes", "subdir1"}, "inc", `inc2 here`},
		{[]string{"layouts"}, "lay", `lay1 here`},
		{[]string{"layouts", "subdir2"}, "lay", `lay2 here`},
		{[]string{"pages"}, "page", `page1 here`},
		{[]string{"pages", "subdir3"}, "page", `page2 here`},
	})
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)

	cfg.Root = dir
	fs := New(cfg)
	err := fs.LookupAll()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("includes: %v\n", fs.IncludeNames())
	fmt.Printf("layouts: %v\n", fs.LayoutNames())
	fmt.Printf("pages: %v\n", fs.PageNames())

	// Output:
	// includes: [inc subdir1/inc]
	// layouts: [lay subdir2/lay]
	// pages: [page subdir3/page]
}

// Here we demonstrate loading a set of templates from a directory.
func ExampleLookupFilesBySuffix() {

	cfg := Config{
		Includes:  ".includes",
		Layouts:   ".layouts",
		Pages:     "not_used",
		Ext:       ".html",
		UseSuffix: true,
	}
	// Here we create a temporary directory and populate it with our sample
	// template definition files; usually the template files would already
	// exist in some location known to the program.
	dir := createTestDir(cfg.Ext, []templateFile{
		{[]string{}, "inc.includes", `inc1 here`},
		{[]string{"subdir1"}, "inc.includes", `inc2 here`},
		{[]string{}, "lay.layouts", `lay1 here`},
		{[]string{"subdir2"}, "lay.layouts", `lay2 here`},
		{[]string{}, "page", `page1 here`},
		{[]string{"subdir3"}, "page", `page2 here`},
	})
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)

	cfg.Root = dir
	fs := New(cfg)
	err := fs.LookupAll()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("includes: %v\n", fs.IncludeNames())
	fmt.Printf("layouts: %v\n", fs.LayoutNames())
	fmt.Printf("pages: %v\n", fs.PageNames())
	// Output:
	// includes: [inc subdir1/inc]
	// layouts: [lay subdir2/lay]
	// pages: [page subdir3/page]
}
