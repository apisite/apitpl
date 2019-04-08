# apitpl
> golang template engine which renders templates by executing them 2 times, one for content and another one for layout

[![GoDoc][gd1]][gd2]
 [![codecov][cc1]][cc2]
 [![Build Status][bs1]][bs2]
 [![GoCard][gc1]][gc2]
 [![GitHub Release][gr1]][gr2]
 ![GitHub code size in bytes][sz]
 [![GitHub license][gl1]][gl2]

[bs1]: https://cloud.drone.io/api/badges/apisite/apitpl/status.svg
[bs2]: https://cloud.drone.io/apisite/apitpl
[cc1]: https://codecov.io/gh/apisite/apitpl/branch/master/graph/badge.svg
[cc2]: https://codecov.io/gh/apisite/apitpl
[gd1]: https://godoc.org/github.com/apisite/apitpl?status.svg
[gd2]: https://godoc.org/github.com/apisite/apitpl
[gc1]: https://goreportcard.com/badge/github.com/apisite/apitpl
[gc2]: https://goreportcard.com/report/github.com/apisite/apitpl
[gr1]: https://img.shields.io/github/release-pre/apisite/apitpl.svg
[gr2]: https://github.com/apisite/apitpl/releases
[sz]: https://img.shields.io/github/languages/code-size/apisite/apitpl.svg
[gl1]: https://img.shields.io/github/license/apisite/apitpl.svg
[gl2]: LICENSE

* Project status: MVP is ready
* Future plans: tests & docs

This package offers 2-step template processing, where page content template called first, so it can
1. change page layout (among them previous markup) and content type
2. abort processing and return error page (this will render layout with error and without content)
3. abort processing and return redirect

If page content template returns HTML, at step 2, layout template will be called for result page markup build.

## Why do we need another template engine?

1. Adding template file without source recompiling
2. Support plain HTML body as template (adding layout without additional markup in content)
3. Attach all (pages,layouts,includes) templates at start (see lookupfs)
4. Auto create routes for all page templates allowing them get required data via api (see ginapitpl)

## Request processing flow

![Request processing flow](flow.png)

## Template structure

As shown in [testdata](tree/master/testdata), site templates tree might looks like:

```
tmpl
├── includes
│   ├── inc.html
│   └── subdir1
│       └── inc.html
├── layouts
│   ├── default.html
│   └── subdir2
│       └── lay.html
└── pages
    ├── page.html
    └── subdir3
        └── page.html
```

## Usage

### Template parsing

All templates from the `Root` directory tree are parsed in `Parse()` call and program should be aborted on error.
Routes for all page URI should be set in `Route()` call after that.

You can enable per request templates parsing for debugging purposes via `ParseAlways(true)` but you still have to restart your program for adding or removing any template file.

### See also
* [Package examples](https://godoc.org/github.com/apisite/apitpl#pkg-examples)
* [ginapitpl](https://github.com/apisite/apitpl/ginapitpl) - [gin](https://github.com/gin-gonic/gin) bindings for this package

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
{{ .Raise 403 true "Error description" }}
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

## See also

* https://stackoverflow.com/questions/42747183/how-to-render-templates-to-multiple-layouts-in-go
* https://medium.com/@leeprovoost/dealing-with-go-template-errors-at-runtime-1b429e8b854a

## License

The MIT License (MIT), see [LICENSE](LICENSE).

Copyright (c) 2018 Aleksei Kovrizhkin <lekovr+apisite@gmail.com>
