package sample

import (
	"github.com/pkg/errors"
	"net/http"
)

// ErrRedirect is an error returned when page needs to be redirected
var ErrRedirect = errors.New("Abort with redirect")

// Meta holds template metadata
type Meta struct {
	Title        string
	contentType  string
	status       int
	error        error
	layout       string
	location     string
	errorMessage string
}

func NewMeta(status int, ctype string) Meta {
	return Meta{status: status, contentType: ctype, layout: "default"}
}

// SetLayout - set page layout
func (p *Meta) SetLayout(name string) (string, error) {
	p.layout = name
	return "", nil
}

// SetTitle - set page title
func (p *Meta) SetTitle(name string) (string, error) {
	p.Title = name
	return "", nil
}

func (p *Meta) SetError(e error) {
	p.error = e
}

// Raise - abort template processing (if given) and raise error
func (p *Meta) Raise(status int, abort bool, message string) (string, error) {
	p.status = status
	p.errorMessage = message // TODO: pass it via error only
	if abort {
		return "", errors.New(message)
	}
	return "", nil
}

// RedirectFound - abort template processing and return redirect with StatusFound status
func (p *Meta) RedirectFound(uri string) (string, error) {
	p.status = http.StatusFound
	p.location = uri
	return "", ErrRedirect // TODO: Is there a way to pass status & title via error?
}

// ErrorMessage returns internal or template error
func (p Meta) ErrorMessage() string {
	if p.errorMessage != "" {
		return p.errorMessage
	}
	if p.error == nil {
		return ""
	}
	return p.error.Error()
}
func (p Meta) Error() error        { return p.error }
func (p Meta) Layout() string      { return p.layout }
func (p Meta) ContentType() string { return p.contentType }
func (p Meta) Status() int         { return p.status }
func (p Meta) Location() string    { return p.location }
