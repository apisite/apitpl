package samplemeta

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewMeta(t *testing.T) {
	m := NewMeta(http.StatusNotImplemented, "text/plain")
	assert.Equal(t, "default", m.Layout())
	assert.Equal(t, http.StatusNotImplemented, m.Status())
}

func TestSetTitle(t *testing.T) {
	m := Meta{}
	m.SetTitle("title")
	assert.Equal(t, "title", m.Title)
}

func TestSetContentType(t *testing.T) {
	m := Meta{}
	m.SetContentType("text/plain")
	assert.Equal(t, "text/plain", m.ContentType())
}

func TestSetLayout(t *testing.T) {
	m := Meta{}
	m.SetLayout("wide")
	assert.Equal(t, "wide", m.Layout())
}

func TestSetStatus(t *testing.T) {
	m := Meta{}
	m.SetStatus(http.StatusNotImplemented)
	assert.Equal(t, http.StatusNotImplemented, m.Status())
}
func TestSetError(t *testing.T) {
	m := Meta{}
	e := errors.New("cause")
	m.SetError(e)
	assert.Equal(t, e, m.Error())
}
