package tpl2x_test

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
