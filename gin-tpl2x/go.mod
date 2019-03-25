module github-my.go/apisite/tpl2x/gin-tpl2x

go 1.12

replace (
	github.com/apisite/tpl2x => ../
	github.com/apisite/tpl2x/gin-tpl2x => ./
)

require (
	github.com/apisite/tpl2x v0.0.0
	github.com/apisite/tpl2x/gin-tpl2x v0.0.0-00010101000000-000000000000
	github.com/birkirb/loggers-mapper-logrus v0.0.0-20180326232643-461f2d8e6f72
	github.com/gin-gonic/gin v1.3.0
	github.com/sirupsen/logrus v1.4.0
	github.com/stretchr/testify v1.3.0
	gopkg.in/birkirb/loggers.v1 v1.1.0
)
