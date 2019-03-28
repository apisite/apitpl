package sample

// Page holds page attributes
type Meta struct {
	Title       string
	contentType string
	status      int
	error       error
	layout      string
}

func NewMeta(status int, ctype string) *Meta {
	return &Meta{status: status, contentType: ctype, layout: "default"}
}

// SetTitle - set page title
func (p *Meta) SetTitle(name string) (string, error) {
	p.Title = name
	return "", nil
}
func (p *Meta) SetError(e error) {
	p.error = e
}

func (p Meta) Error() error {
	return p.error
}

func (p Meta) Layout() string {
	return p.layout
}

func (p Meta) ContentType() string { return p.contentType }
func (p Meta) Status() int         { return p.status }
