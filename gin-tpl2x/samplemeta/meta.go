// Package samplemeta implements sample type Meta which holds template metadata
package samplemeta

import (
	"github.com/pkg/errors"
	"net/http"

	base "github.com/apisite/tpl2x/samplemeta"
)

// ErrRedirect is an error returned when page needs to be redirected
var ErrRedirect = errors.New("Abort with redirect")

// Meta holds template metadata
type Meta struct {
	base.Meta
	location string
	// store original message because error stack may change it
	errorMessage string
}

// NewMeta returns new initialised Meta struct
func NewMeta(status int, ctype string) *Meta {
	m := base.NewMeta(status, ctype)
	return &Meta{Meta: *m}
}

// Raise interrupts template processing (if given) and raise error
func (p *Meta) Raise(status int, abort bool, message string) (string, error) {
	p.SetStatus(status)
	p.errorMessage = message // TODO: pass it via error only
	if abort {
		return "", errors.New(message)
	}
	return "", nil
}

// RedirectFound interrupts template processing and return redirect with StatusFound status
func (p *Meta) RedirectFound(uri string) (string, error) {
	p.SetStatus(http.StatusFound)
	p.location = uri
	return "", ErrRedirect // TODO: Is there a way to pass status & title via error?
}

// ErrorMessage returns internal or template error message
func (p Meta) ErrorMessage() string {
	if p.errorMessage != "" {
		return p.errorMessage
	}
	if p.Error() == nil {
		return ""
	}
	return p.Error().Error()
}

// Location returns redirect location
func (p Meta) Location() string { return p.location }
