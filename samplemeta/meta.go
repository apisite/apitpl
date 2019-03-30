// Package samplemeta implements sample type Meta which holds template metadata
package samplemeta

// Meta holds template metadata
type Meta struct {
	Title  string
	error  error
	layout string

	// Used in http test
	contentType string
	status      int
}

// NewMeta returns new initialised Meta struct
func NewMeta(status int, ctype string) *Meta {
	return &Meta{status: status, contentType: ctype, layout: "default"}
}

// SetTitle sets page title
func (m *Meta) SetTitle(name string) string { m.Title = name; return "" }

// SetLayout sets page layout
func (m *Meta) SetLayout(name string) string { m.layout = name; return "" }

// SetStatus sets response status
func (m *Meta) SetStatus(status int) string { m.status = status; return "" }

// Layout returns page layout
func (m Meta) Layout() string { return m.layout }

// SetError sets error by template engine
// Not for use in templates (see Raise)
func (m *Meta) SetError(e error) { m.error = e }

// Error returns page error
func (m Meta) Error() error { return m.error }

// ContentType returns page content type
func (m Meta) ContentType() string { return m.contentType }

// Status returns page status
func (m Meta) Status() int { return m.status }
