module github.com/apisite/apitpl

go 1.12

require (
	github.com/birkirb/loggers-mapper-logrus v0.0.0-20180326232643-461f2d8e6f72
	github.com/blang/vfs v1.0.0 // indirect
	github.com/daaku/go.zipexe v1.0.0 // indirect
	github.com/gin-gonic/gin v1.5.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190227141107-8c4636f812cc
	github.com/phogolabs/parcello v0.8.1
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20190326090315-15845e8f865b // indirect
	gopkg.in/birkirb/loggers.v1 v1.1.0
)

replace github.com/apisite/apitpl/ginapitpl => ./ginapitpl
