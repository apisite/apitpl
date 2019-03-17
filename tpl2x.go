// Package tpl2x renders templates by execiting them twice, one for content and one for layout
package tpl2x

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"html/template"
	"io"

	"github.com/oxtoacart/bpool"

	"github.com/apisite/tpl2x/lookupfs"
)

// Config holds config variables and its defaults
type Config struct {
	ContentType string `long:"content-type" default:"text/html; charset=utf-8" description:"Default content type"`
	BufferSize  int    `long:"buffer" default:"64" description:"Template buffer size"`
}

type TemplateService struct {
	config           Config
	lfs              *lookupfs.LookupFileSystem
	funcMap          template.FuncMap // used externally as default runtime func map
	layouts          *map[string]*template.Template
	pages            *map[string]*template.Template
	baseTemplate     *template.Template
	bufPool          *bpool.BufferPool
	useCustomContent bool
}

func New(cfg Config) (tfs *TemplateService) {
	tfs = &TemplateService{
		config: cfg,
		funcMap: template.FuncMap{
			"content": func() string { return "" },
		},
		bufPool: bpool.NewBufferPool(cfg.BufferSize),
	}
	return tfs
}

func (tfs *TemplateService) LookupFS(fs *lookupfs.LookupFileSystem) *TemplateService {
	tfs.lfs = fs
	return tfs
}

func (tfs *TemplateService) Funcs(funcMap template.FuncMap) *TemplateService {
	for k, v := range funcMap {
		tfs.funcMap[k] = v
		if k == "content" {
			tfs.useCustomContent = true
		}
	}
	return tfs
}

func (tfs *TemplateService) PageNames() []string {
	return tfs.lfs.PageNames()
}
func (tfs *TemplateService) Parse() (*TemplateService, error) {

	err := tfs.lfs.LookupAll()
	if err != nil {
		return nil, err
	}

	includes, err := tfs.ParseIncludes(tfs.lfs.Includes)
	if err != nil {
		return nil, err
	}
	tfs.baseTemplate = includes

	layouts, err := tfs.ParseTemplates(tfs.lfs.Layouts)
	if err != nil {
		return nil, err
	}

	pages, err := tfs.ParseTemplates(tfs.lfs.Pages)
	if err != nil {
		return nil, err
	}

	tfs.layouts = layouts
	tfs.pages = pages
	return tfs, nil
}

func (tfs TemplateService) ParseIncludes(items map[string]lookupfs.File) (*template.Template, error) {
	var t *template.Template

	for k, f := range items {
		s, err := tfs.lfs.ReadFile(f.Path)
		if err != nil {
			return nil, err
		}
		var tmpl *template.Template
		if t == nil {
			t = template.New(k)
			tmpl = t
		} else {
			tmpl = t.New(k)
		}
		//		fmt.Printf("inc (%s): %s\n", k, s)
		_, err = tmpl.Funcs(tfs.funcMap).Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (tfs TemplateService) ParseTemplates(items map[string]lookupfs.File) (*map[string]*template.Template, error) {
	templates := map[string]*template.Template{}
	for k, f := range items {
		s, err := tfs.lfs.ReadFile(f.Path)
		if err != nil {
			return nil, err
		}
		t := tfs.baseTemplate
		//		fmt.Printf("tmpl (%s): %s\n", k, s)
		var tmpl *template.Template
		if t == nil {
			tmpl = template.New(k)
		} else {
			tmpl, err = t.Clone()
			if err != nil {
				return nil, err
			}
			tmpl = tmpl.New(k)
		}
		_, err = tmpl.Funcs(tfs.funcMap).Parse(s)
		if err != nil {
			return nil, errors.Wrap(err, "process layout")
		}
		templates[k] = tmpl
	}
	return &templates, nil
}
func (tfs TemplateService) Execute(wr io.Writer, name string, funcs template.FuncMap, data interface{}) error {
	content, err := tfs.RenderContent(name, funcs, data)
	if err != nil {
		return err
	}
	//defer tfs.bufPool.Put(content)
	return tfs.Render(wr, funcs, data, content)
}

func (tfs TemplateService) RenderContent(name string, funcs template.FuncMap, data interface{}) (*bytes.Buffer, error) {
	tmpl, ok := (*tfs.pages)[name] // TODO: tfs.Lookup(tfs.pages, name)
	if !ok {
		err := fmt.Errorf("The page %s does not exist.", name)
		//p.Raise(http.StatusNotFound, "NOT FOUND", err.Error(), false)
		return nil, err
	}
	//	p.uri = name
	//	p.ContentType = tfs.config.ContentType
	//	p.Status = http.StatusOK
	buf := tfs.bufPool.Get()
	err := tmpl.Funcs(funcs).ExecuteTemplate(buf, name, data)
	if err != nil {
		tfs.bufPool.Put(buf)
		return nil, errors.Wrap(err, "exec page")
	}
	//	fmt.Printf("content (%s): %s\n", uri, buf)

	return buf, nil //template.HTML(buf.Bytes())
	//tfs.bufPool.Put(buf)

}

// RenderLayout renders page content in given layout
func (tfs TemplateService) Render(w io.Writer, funcs template.FuncMap, data interface{}, content *bytes.Buffer) (err error) {
	name := tfs.lfs.DefaultLayout()
	tmpl, ok := (*tfs.layouts)[name] // TODO: tfs.Lookup(tfs.pages, name)
	if !ok {
		err := fmt.Errorf("layout %s does not exist.", name)
		//p.Raise(http.StatusNotFound, "NOT FOUND", err.Error(), false)
		return err
	}
	buf := tfs.bufPool.Get()
	if !tfs.useCustomContent {
		funcs["content"] = func() string { return content.String() } // template.HTML(content.Bytes())
	}

	err = tmpl.Funcs(funcs).ExecuteTemplate(buf, name, data)
	if content != nil {
		tfs.bufPool.Put(content)

	}

	if err != nil {
		return errors.Wrap(err, "exec layout1")
	}

	buf.WriteTo(w)
	tfs.bufPool.Put(buf)
	return nil
}
