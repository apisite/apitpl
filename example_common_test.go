package tpl2x_test

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// BufferPool size for rendered templates
const bufferSize int = 64

// Page holds page attributes
type Meta struct {
	Title       string
	ContentType string
	Status      int
	error       error
	layout      string
}

// SetTitle - set page title
func (p *Meta) SetTitle(name string) (string, error) {
	p.Title = name
	return "", nil
}
func (p *Meta) SetError(e error) {
	p.error = e
}

func (p Meta) Error() error {
	return p.error
}

func (p Meta) Layout() string {
	return "default"
}

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
