package lookupfs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewErrors(t *testing.T) {
	cfg := Config{
		Ext: ".html",
	}
	dir := createTestDir(cfg.Ext, []templateFile{
		{[]string{"includes"}, "inc", `inc1 here`},
		{[]string{"layouts"}, "lay", `lay1 here`},
		{[]string{"pages"}, "page", `page1 here`},
	})
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)

	tests := []struct {
		name string
		cfg  Config
		err  string
	}{
		{name: "PrefixNoIncludesDir", cfg: Config{Includes: "404"}, err: "includes walk failed: lstat"},
		{name: "PrefixNoLayoutDir", cfg: Config{Includes: "includes", Layouts: "404"}, err: "layouts walk failed: lstat"},
		{name: "PrefixNoPageDir", cfg: Config{Includes: "includes", Layouts: "layouts", Pages: "404"}, err: "pages walk failed: lstat"},
		{name: "SuffixNoFiles", cfg: Config{UseSuffix: true, Root: filepath.Join(dir, "404")}, err: "walk failed: walk error: lstat"},
	}
	for _, tt := range tests {
		if tt.cfg.Root == "" {
			tt.cfg.Root = dir

		}
		err := New(tt.cfg).LookupAll()
		require.NotNil(t, err)
		//fmt.Println(errors.Cause(err))
		assert.True(t, strings.HasPrefix(err.Error(), tt.err), err.Error())
	}
}
