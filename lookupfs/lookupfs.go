// Package lookupfs implements a filesystem backend for tpl2x.
// It can use native filesystem (by default) or embedded filesystem (which can be set via FileSystem func).
package lookupfs

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Config holds config variables and its defaults
type Config struct {
	Root      string `long:"templates" default:"tmpl/" description:"Templates root path"`
	Ext       string `long:"mask" default:".tmpl" description:"Templates filename mask"`
	Includes  string `long:"includes" default:"inc/" description:"Includes path"`
	Layouts   string `long:"layouts" default:"layout/" description:"Layouts path"`
	Pages     string `long:"pages" default:"page/" description:"Pages path"`
	UseSuffix bool   `long:"use_suffix" description:"Template type defined by suffix"`
	Index     string `long:"index" default:"index" description:"Index page name"`
	DefLayout string `long:"def_layout" default:"default" description:"Default layout template"`
}

type defaultFS struct{}

func (fs defaultFS) Walk(path string, wf filepath.WalkFunc) error {
	return filepath.Walk(path, wf)
}

func (fs defaultFS) Open(name string) (http.File, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err // TODO: What with mapDirOpenError(err, fullName)?
	}
	return f, nil
}

// FileSystem holds all of used filesystem access methods
type FileSystem interface {
	Walk(root string, walkFn filepath.WalkFunc) error
	Open(name string) (http.File, error)
}

// File holds file metadata
type File struct {
	Path    string
	ModTime time.Time
}

// LookupFileSystem holds filesystem with template lookup functionality
type LookupFileSystem struct {
	config       Config
	fs           FileSystem
	Includes     map[string]File
	Layouts      map[string]File
	Pages        map[string]File
	disableCache bool
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

/*
// DisableCache disables template caching
func (lfs *LookupFileSystem) DisableCache(flag bool) {
	lfs.disableCache = flag
}
*/

// IncludeNames return sorted slice of include names
func (lfs LookupFileSystem) IncludeNames() []string {
	return mapKeys(lfs.Includes)
}

// LayoutNames return sorted slice of layout names
func (lfs LookupFileSystem) LayoutNames() []string {
	return mapKeys(lfs.Layouts)
}

// PageNames return sorted slice of page names
func (lfs LookupFileSystem) PageNames() []string {
	return mapKeys(lfs.Pages)
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

func (lfs LookupFileSystem) walk(tag, prefix string, files *map[string]File) (err error) {

	root := filepath.Join(lfs.config.Root, prefix)
	err = lfs.fs.Walk(root, func(path string, f os.FileInfo, err error) error {
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

		// Do not begin with a slash
		if name != "/" {
			name = strings.TrimPrefix(name, "/")
		}

		//fmt.Printf("Found %s -> %s\n", name, path)
		(*files)[name] = File{Path: path, ModTime: f.ModTime()}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, tag+" walk failed")
	}
	return nil
}

func (lfs *LookupFileSystem) lookupFilesByPrefix() (err error) {

	if err = lfs.walk("includes", lfs.config.Includes, &lfs.Includes); err != nil {
		return
	}
	if err = lfs.walk("layouts", lfs.config.Layouts, &lfs.Layouts); err != nil {
		return
	}
	if err = lfs.walk("pages", lfs.config.Pages, &lfs.Pages); err != nil {
		return
	}

	return
}

func (lfs *LookupFileSystem) lookupFilesBySuffix() (err error) {

	err = lfs.fs.Walk(lfs.config.Root, func(path string, f os.FileInfo, err error) error {
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

		// Do not begin with a slash
		name = strings.TrimPrefix(name, "/")

		value := File{Path: path, ModTime: f.ModTime()}
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
func mapKeys(m map[string]File) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.StringSlice(keys).Sort()
	return keys
}
