// Package lookupfs implements a filesystem backend for apitpl.
// It can use native filesystem (by default) or embedded filesystem (which can be set via FileSystem func).
package lookupfs

import (
	"github.com/pkg/errors"
	"io/fs"
	"io/ioutil"
//	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// codebeat:disable[TOO_MANY_IVARS]

// Config holds config variables and its defaults
type Config struct {
	Root       string `long:"templates" default:"tmpl/" description:"Templates root path"`
	Ext        string `long:"mask" default:".tmpl" description:"Templates filename mask"`
	Includes   string `long:"includes" default:"inc/" description:"Includes path"`
	Layouts    string `long:"layouts" default:"layout/" description:"Layouts path"`
	Pages      string `long:"pages" default:"page/" description:"Pages path"`
	UseSuffix  bool   `long:"use_suffix" description:"Template type defined by suffix"`
	Index      string `long:"index" default:"index" description:"Index page name"`
	DefLayout  string `long:"def_layout" default:"default" description:"Default layout template"`
	HidePrefix string `long:"hide_prefix" default:"." description:"Treat files with this prefix as hidden"`
}

// codebeat:enable[TOO_MANY_IVARS]

type defaultFS struct{}

func (dfs defaultFS) Walk(path string, wf fs.WalkDirFunc) error {
	return fs.WalkDir(dfs, path, wf)
}

func (dfs defaultFS) Open(name string) (fs.File, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err // TODO: What with mapDirOpenError(err, fullName)?
	}
	return f, nil
}

// FileSystem holds all of used filesystem access methods
type FileSystem interface {
//	Walk(root string, walkFn filepath.WalkFunc) error
	Open(name string) (fs.File, error)
}

// File holds file metadata
type File struct {
	Path    string
	ModTime time.Time
}

// LookupFileSystem holds filesystem with template lookup functionality
type LookupFileSystem struct {
	config   Config
	fs       FileSystem
	Includes map[string]File
	Layouts  map[string]File
	Pages    map[string]File
}

// New creates LookupFileSystem
func New(cfg Config) *LookupFileSystem {
	return &LookupFileSystem{
		config:   cfg,
		fs:       defaultFS{},
		Includes: map[string]File{},
		Layouts:  map[string]File{},
		Pages:    map[string]File{},
	}
}

// FileSystem changes filesystem access object
func (lfs *LookupFileSystem) FileSystem(fs FileSystem) *LookupFileSystem {
	lfs.fs = fs
	return lfs
}

// DefaultLayout returns default layout name
// This name has been checked for availability in LookupAll()
func (lfs LookupFileSystem) DefaultLayout() string {
	return lfs.config.DefLayout
}

// IncludeNames return sorted slice of include names
func (lfs LookupFileSystem) IncludeNames() []string {
	return mapKeys(lfs.Includes, "", false)
}

// LayoutNames return sorted slice of layout names
func (lfs LookupFileSystem) LayoutNames() []string {
	return mapKeys(lfs.Layouts, "", false)
}

// PageNames return sorted slice of page names
func (lfs LookupFileSystem) PageNames(hide bool) []string {
	return mapKeys(lfs.Pages, lfs.config.HidePrefix, hide)
}

// LookupAll scan filesystem for includes,pages and layouts
func (lfs *LookupFileSystem) LookupAll() (err error) {
	if lfs.config.UseSuffix {
		err = lfs.lookupFilesBySuffix()
	} else {
		err = lfs.lookupFilesByPrefix()
	}
	if err == nil {
		if _, ok := lfs.Layouts[lfs.DefaultLayout()]; !ok {
			err = errors.Errorf("default layout (%s) does not exists", lfs.DefaultLayout())
		}
	}
	return
}

// ReadFile reads file via filesystem method
func (lfs LookupFileSystem) ReadFile(name string) (string, error) {
	f, err := lfs.fs.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	s := string(b)
	return s, nil
}

func (lfs LookupFileSystem) walk(tag, prefix string, files map[string]File) (err error) {

	root := filepath.Join(lfs.config.Root, prefix)
	err = fs.WalkDir(lfs.fs, root, func(path string, f fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		// Remove root prefix and ext suffix
		name := strings.TrimPrefix(strings.TrimSuffix(path, lfs.config.Ext), root)

		// Convert filepath to uri if system is non-POSIX
		name = filepath.ToSlash(name)

		// Do not end with an index
		name = strings.TrimSuffix(name, lfs.config.Index)

		// Replace /__ with /: (':' used for params in gin)
		name = strings.ReplaceAll(name, "/__", "/:")

		// Do not begin with a slash
		if name != "/" {
			name = strings.TrimPrefix(name, "/")
		}

		//fmt.Printf("Found %s -> %s\n", name, path)
		info,_ := f.Info()
		files[name] = File{Path: path, ModTime: info.ModTime()}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, tag+" walk failed")
	}
	return nil
}

func (lfs *LookupFileSystem) lookupFilesByPrefix() (err error) {

	if lfs.config.Includes != "" {
		if err = lfs.walk("includes", lfs.config.Includes, lfs.Includes); err != nil {
			return
		}
	}
	if err = lfs.walk("layouts", lfs.config.Layouts, lfs.Layouts); err != nil {
		return
	}
	if err = lfs.walk("pages", lfs.config.Pages, lfs.Pages); err != nil {
		return
	}

	return
}

func (lfs *LookupFileSystem) lookupFilesBySuffix() (err error) {

	err = fs.WalkDir(lfs.fs, lfs.config.Root, func(path string, f fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "walk error")
		}
		if f.IsDir() {
			return nil
		}

		// Remove root prefix and ext suffix
		name := strings.TrimPrefix(strings.TrimSuffix(path, lfs.config.Ext), lfs.config.Root)

		// Convert filepath to uri if system is non-POSIX
		name = filepath.ToSlash(name)

		// Replace /__ with /: (':' used for params in gin)
		name = strings.ReplaceAll(name, "/__", "/:")

		// Do not begin with a slash
		if name != "/" {
			name = strings.TrimPrefix(name, "/")
		}

		info,_ := f.Info()
		value := File{Path: path, ModTime: info.ModTime()}
		if strings.HasSuffix(name, lfs.config.Includes) {
			lfs.Includes[strings.TrimSuffix(name, lfs.config.Includes)] = value
		} else if strings.HasSuffix(name, lfs.config.Layouts) {
			lfs.Layouts[strings.TrimSuffix(name, lfs.config.Layouts)] = value
		} else {
			// only page templates must be here
			// no suffixes => no checking
			lfs.Pages[name] = value
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "walk failed")
	}
	return nil
}

// mapKeys returns sorted map keys
func mapKeys(m map[string]File, prefix string, hide bool) []string {
	var keys []string // len depends on hide
	for k := range m {
		if hide && (strings.HasPrefix(k, prefix) || strings.Contains(k, "/"+prefix)) {
			// Skip hidden pages
			continue
		}
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()
	return keys
}
