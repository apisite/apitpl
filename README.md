# tpl2x
> golang templates executed twice, for content and for layout

[![GoDoc][gd1]][gd2]
 [![codecov][cc1]][cc2]
 [![GoCard][gc1]][gc2]
 [![GitHub Release][gr1]][gr2]
 [![GitHub code size in bytes][sz]]()
 [![GitHub license][gl1]][gl2]

[cc1]: https://codecov.io/gh/apisite/tpl2x/branch/master/graph/badge.svg
[cc2]: https://codecov.io/gh/apisite/tpl2x
[gd1]: https://godoc.org/github.com/apisite/tpl2x?status.svg
[gd2]: https://godoc.org/github.com/apisite/tpl2x
[gc1]: https://goreportcard.com/badge/github.com/apisite/tpl2x
[gc2]: https://goreportcard.com/report/github.com/apisite/tpl2x
[gr1]: https://img.shields.io/github/release-pre/apisite/tpl2x.svg
[gr2]: https://github.com/apisite/tpl2x/releases
[sz]: https://img.shields.io/github/languages/code-size/apisite/tpl2x.svg
[gl1]: https://img.shields.io/github/license/apisite/tpl2x.svg
[gl2]: LICENSE

* Project status: MVP is ready
* Future plans: tests & docs

This package offers 2-step template processing, where page content template called first, so it can
1. change page layout (among them previous markup)
2. abort processing and return error page (this will go to way 1)
3. abort processing and return redirect

If page content template returns HTML, at step 2, layout template will be called for result page markup build.

## Why do we need another template engine?

1. Adding template file without source recompiling
2. Support plain HTML body as template (adding layout without additional markup)

## Request processing flow

![Request processing flow](flow.png)

## Template structure

As shown in examples, site templates tree might looks like:

```
tmpl
├── inc
│   ├── footer.tmpl
│   ├── header.tmpl
│   └── menu.tmpl
├── layout
│   ├── default.tmpl
│   ├── authorized.tmpl
│   └── wide.tmpl
└── page
    ├── admin
    │   └── index.tmpl
    ├── index.tmpl
    └── page.tmpl

```

## Usage

### See also
* [Package examples](https://godoc.org/github.com/apisite/tpl2x#pkg-examples)
* [gin-tpl2x](https://github.com/apisite/tpl2x/gin-tpl2x) - [gin](https://github.com/gin-gonic/gin) bindings for this package

### Template methods
Get http.Request data
```
{{ request.Host }}{{ request.URL.String | HTML }}
```
Get query params
```
{{ $param := .Request.URL.Query.Get "param" -}}
```
Set page title
```
{{ .SetTitle "admin:index" -}}
```
Choose layout
```
{{ .SetLayout "wide" -}}
```
Stop template processing and raise error
```
{{ .Raise 403 "Test error" "Error description" true }}
```
Stop template processing and return redirect 
```
{{ .RedirectFound "/page" }}
```

### Custom methods
in code
```go
reqFuncs["data"] = func() interface{} {
    return data
}
p, err := mlt.RenderPage(uri, reqFuncs, r)
```
in templates
```
{{range data.Todos -}}
    <li>{{- .Title }}
{{end -}}

```

## License

The MIT License (MIT), see [LICENSE](LICENSE).

Copyright (c) 2018 Aleksei Kovrizhkin <lekovr+apisite@gmail.com>
