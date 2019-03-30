package tpl2x

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/apisite/tpl2x/lookupfs"
	"github.com/apisite/tpl2x/samplemeta"
)

type ServerSuite struct {
	suite.Suite
	cfg lookupfs.Config
	srv *TemplateService
}

func TestSuite(t *testing.T) {
	myTest := &ServerSuite{}
	suite.Run(t, myTest)
}

func (ss *ServerSuite) SetupSuite() {

	// BufferPool size for rendered templates
	const bufferSize int = 64

	ss.cfg = lookupfs.Config{
		Layouts:   "layouts",
		Pages:     "pages",
		Ext:       ".html",
		DefLayout: "default",
		Root:      "testdata",
	}

	tfs, err := New(bufferSize).
		LookupFS(lookupfs.New(ss.cfg)).
		ParseAlways(true).
		Parse()
	require.NoError(ss.T(), err)
	ss.srv = tfs
}

func (ss *ServerSuite) TestNoIncludes() {
	page := &samplemeta.Meta{}
	var b bytes.Buffer
	err := ss.srv.Execute(&b, "page", template.FuncMap{}, page)
	require.NoError(ss.T(), err)
}

func (ss *ServerSuite) TestPageNotExists() {
	page := &samplemeta.Meta{}
	page.SetLayout("simple")
	var b bytes.Buffer
	err := ss.srv.Execute(&b, "page_unknown", template.FuncMap{}, page)
	require.NoError(ss.T(), err)
	assert.Equal(ss.T(), "<title>Error 0: Sorry</title>\nThe page page_unknown does not exist.\n", b.String())
}
func (ss *ServerSuite) TestLayoutNotExists() {
	page := &samplemeta.Meta{}
	page.SetLayout("unknown")
	var b bytes.Buffer
	err := ss.srv.Execute(&b, "page", template.FuncMap{}, page)
	assert.Equal(ss.T(), "exec layout: html/template: \"unknown\" is undefined", err.Error())
}
