module github.com/apisite/tpl2x

go 1.12

replace (
	github.com/apisite/tpl2x => ./
	github.com/apisite/tpl2x/gin-tpl2x => ./gin-tpl2x
)

require (
	github.com/apisite/tpl2x/gin-tpl2x v0.0.0-00010101000000-000000000000 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190227141107-8c4636f812cc
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.3.0
)
