// Package tpl2x implements template engine
// which renders templates by executing them 2 times, one for content and another one for layout.
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

// TemplateService holds templates data & methods
type TemplateService struct {
	lfs              *lookupfs.LookupFileSystem
	funcMap          template.FuncMap
	layouts          *map[string]*template.Template
	pages            *map[string]*template.Template
	baseTemplate     *template.Template
	bufPool          *bpool.BufferPool
	useCustomContent bool
	parseAlways      bool
}

// MetaData holds template metadata access methods
type MetaData interface {
	Error() error   // Template exec error
	SetError(error) // Store unhandler template error
	Layout() string // Return layout name
}

// New creates TemplateService with BufferPool of given size
func New(size int) (tfs *TemplateService) {
	tfs = &TemplateService{
		funcMap: template.FuncMap{
			"content": func() string { return "" },
		},
		bufPool: bpool.NewBufferPool(size),
	}
	return tfs
}

// LookupFS sets lookup filesystem
func (tfs *TemplateService) LookupFS(fs *lookupfs.LookupFileSystem) *TemplateService {
	tfs.lfs = fs
	return tfs
}

// ParseAlways disables template caching
func (tfs *TemplateService) ParseAlways(flag bool) *TemplateService {
	tfs.parseAlways = flag
	return tfs
}

// Funcs loads initial funcmap
func (tfs *TemplateService) Funcs(funcMap template.FuncMap) *TemplateService {
	for k, v := range funcMap {
		tfs.funcMap[k] = v
		if k == "content" {
			tfs.useCustomContent = true
		}
	}
	return tfs
}

// PageNames returns page names for router setup
func (tfs TemplateService) PageNames(hide bool) []string {
	return tfs.lfs.PageNames(hide)
}

// Parse parses all of service templates
func (tfs *TemplateService) Parse() (*TemplateService, error) {

	err := tfs.lfs.LookupAll()
	if err != nil {
		return nil, err
	}

	includes, err := tfs.parseIncludes(tfs.lfs.Includes)
	if err != nil {
		return nil, err
	}

	layouts, err := tfs.parseTemplates(includes, tfs.lfs.Layouts)
	if err != nil {
		return nil, err
	}

	pages, err := tfs.parseTemplates(includes, tfs.lfs.Pages)
	if err != nil {
		return nil, err
	}

	tfs.baseTemplate = includes
	tfs.layouts = layouts
	tfs.pages = pages
	return tfs, nil
}

// parseIncludes parses included templates
func (tfs TemplateService) parseIncludes(items map[string]lookupfs.File) (*template.Template, error) {
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
		_, err = tmpl.Funcs(tfs.funcMap).Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

// parseTemplates parses all page & layout templates
func (tfs TemplateService) parseTemplates(includes *template.Template, items map[string]lookupfs.File) (*map[string]*template.Template, error) {
	templates := map[string]*template.Template{}
	for k, f := range items {
		tmpl, err := tfs.parseTemplate(includes, k, f)
		if err != nil {
			return nil, errors.Wrap(err, "parse template")
		}
		templates[k] = tmpl
	}
	return &templates, nil
}

// parseTemplate parses single template
func (tfs TemplateService) parseTemplate(includes *template.Template, k string, f lookupfs.File) (*template.Template, error) {
	s, err := tfs.lfs.ReadFile(f.Path)
	if err != nil {
		return nil, err
	}
	t := includes
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
	return tmpl, err
}

func (tfs TemplateService) parseTemplateWithDeps(items map[string]lookupfs.File, name string) (*template.Template, error) {
	includes, err := tfs.parseIncludes(tfs.lfs.Includes)
	if err != nil {
		return nil, err
	}
	f, ok := items[name]
	if !ok {
		err := fmt.Errorf("The page %s does not exist.", name)
		return nil, err
	}
	return tfs.parseTemplate(includes, name, f)
}

// Execute renders page content and layout
func (tfs TemplateService) Execute(wr io.Writer, name string, funcs template.FuncMap, data MetaData) error {
	return tfs.Render(wr, funcs, data, tfs.RenderContent(name, funcs, data))
}

// RenderContent renders page content
func (tfs TemplateService) RenderContent(name string, funcs template.FuncMap, data MetaData) *bytes.Buffer {
	var tmpl *template.Template
	var err error
	if tfs.parseAlways {
		tmpl, err = tfs.parseTemplateWithDeps(tfs.lfs.Pages, name)
		if err != nil {
			data.SetError(err)
			return nil
		}
	} else {
		var ok bool
		tmpl, ok = (*tfs.pages)[name] // TODO: tfs.Lookup(tfs.pages, name)
		if !ok {
			err = fmt.Errorf("The page %s does not exist.", name)
			data.SetError(err)
			return nil
		}
	}
	buf := tfs.bufPool.Get()
	err = tmpl.Funcs(funcs).ExecuteTemplate(buf, name, data)
	if err != nil {
		tfs.bufPool.Put(buf)
		data.SetError(err)
		return nil
	}
	return buf
}

// layout returns metadata layout (if exists) or default layout otherwise
func (tfs TemplateService) layout(name string, data MetaData) *template.Template {
	tmpl, ok := (*tfs.layouts)[name]
	if !ok {
		err := fmt.Errorf("layout %s does not exist", name)
		data.SetError(err)
		tmpl = (*tfs.layouts)[tfs.lfs.DefaultLayout()]
	}
	return tmpl
}

// Render renders layout with prepared content
func (tfs TemplateService) Render(w io.Writer, funcs template.FuncMap, data MetaData, content *bytes.Buffer) (err error) {

	name := data.Layout()
	if name == "" {
		// No layout needed
		if content != nil {
			content.WriteTo(w)
			tfs.bufPool.Put(content)
		}
		return nil
	}
	var tmpl *template.Template
	if tfs.parseAlways {
		var err error
		tmpl, err = tfs.parseTemplateWithDeps(tfs.lfs.Layouts, name)
		if err != nil {
			data.SetError(err)
			// TODO: parse default layout?
			tmpl = (*tfs.layouts)[tfs.lfs.DefaultLayout()]
		}
	} else {
		tmpl = tfs.layout(name, data)
	}
	buf := tfs.bufPool.Get()
	defer tfs.bufPool.Put(buf)
	if !tfs.useCustomContent && content != nil {
		funcs["content"] = func() string { return content.String() }
	}
	err = tmpl.Funcs(funcs).ExecuteTemplate(buf, name, data)
	if content != nil {
		tfs.bufPool.Put(content)
	}
	if err != nil {
		return errors.Wrap(err, "exec layout")
	}
	buf.WriteTo(w)
	return nil
}
