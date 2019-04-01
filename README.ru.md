# apisite/apitpl
> Рендеринг дерева шаблонов, использующих API

[![GoDoc][gd1]][gd2]
 [![codecov][cc1]][cc2]
 [![Build Status][bs1]][bs2]
 [![GoCard][gc1]][gc2]
 [![GitHub Release][gr1]][gr2]
 ![GitHub code size in bytes][sz]
 [![GitHub license][gl1]][gl2]

[bs1]: https://cloud.drone.io/api/badges/apisite/tpl2x/status.svg
[bs2]: https://cloud.drone.io/apisite/tpl2x
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

* Готовность к релизу 1.0: 95%
* Задачи до релиза:
  * [ ] актуализировать документацию
  * [ ] определиться с именем репозитория (apitpl)
  * [ ] решить вопрос хранения имени метода `content` 

## Введение

В проекте [apisite](https://github.com/apisite/apisite) шаблоны оперируют только данными HTTP-запроса и результатами вызовов методов API. В следствие того, что для страниц не требуется индивидуальная подготовка данных, необходим универсальный роутер между адресами страниц и их шаблонами. Библиотека [apitpl](https://github.com/apisite/tpl2x) посвящена решению этой задачи.

_Примечания._

В данном документе
* используется термин "tpl2x" - предыдущее название библиотеки apitpl
* термином "шаблон" обозначен файл в синтаксисе [html/template](https://golang.org/pkg/html/template/)

## Исходные данные

Библиотека работает с деревом каталогов, содержащих шаблоны трех типов:

* страницы (pages) - шаблоны, формирующие контент страницы (по пути к файлу формируется URL страницы)
* макеты (layouts) - шаблоны, состоящие из общих для всех страниц элементов (шапки, подвала и т.п.), в которые помещается контент (кроме этой вставки, шаблон может быть обычной HTML-страницей)
* включения (includes) - блоки, которые могут использоваться в страницах и макетах

См. также: [Пример дерева шаблонов](https://github.com/apisite/tpl2x/tree/master/gin-tpl2x/testdata)

## Возможности

* разделение шаблонов по типам на основе префикса или суффикса имени файла
* поддержка работы с шаблонами из встроенной файловой системы
* рендеринг страниц и макетов с использованием инклюдов, с однократным парсингом и парсингом по запросу (флаг `ParseAlways(true)`)
* [роутинг для net/http](https://godoc.org/github.com/apisite/tpl2x#example-package--Http) и [роутинг для gin-gonic/gin](https://godoc.org/github.com/apisite/tpl2x/gin-tpl2x#example-package)

## Структура

Библиотека разделена на следующие части:

* [tpl2x](https://godoc.org/github.com/apisite/tpl2x) - код формирования страницы из шаблона (выполняется в два этапа - формирование контента и сборка страницы по макеты с включением в него контента)
* [lookupfs](https://godoc.org/github.com/apisite/tpl2x/lookupfs) - получение из файловой системы (обычной или встроенной) списков шаблонов
* [samplemeta](https://godoc.org/github.com/apisite/tpl2x/samplemeta) - пример метаданных, которые могут передаваться из шаблона контента в шаблон макета
* [samplefs](https://godoc.org/github.com/apisite/tpl2x/samplefs) - встроенная файловая система для сокращения кода примеров, используется только в `example_*_test.go` 
* [gin-tpl2x](https://godoc.org/github.com/apisite/tpl2x/gin-tpl2x) - интеграция функционала tpl2x в [gin](https://github.com/gin-gonic/gin) (код оформлен модулем, чтобы его зависимости не попали в остальные части), для тестов и примеров этот модуль имеет свои копии [samplemeta](https://godoc.org/github.com/apisite/tpl2x/gin-tpl2x/samplemeta) и [samplefs](https://godoc.org/github.com/apisite/tpl2x/gin-tpl2x/samplefs)

## Особенности реализации

### tpl2x 

Формирование страницы из шаблона выполняется в два шага:
1. формирование контента 
2. формирование страницы по макету с использованием контента

Такое разделение преследует следующие цели:

* процесс формирования контента может быть прерван вызовом исключения и при работе с макетом этот контент не будет использован (например, будет выведено описание этого исключения)
* процесс формирования контента может быть прерван с возвратом переадресации, в этом случае макет не будет использован и ответом сервера станет переадресация
* при формировании контента можно в любой момент изменить макет, который будет использован на шаге 2 

При этом на сам шаблон страницы не накладывается никаких ограничений, это может быть как golang template так и блок HTML разметки. Потенциально, это позволит включать шаблоны страниц в другие страницы (если это понадобится, но такое еще не реализовано).

### lookupfs

Получение списка файлов для каждого из трех типов шаблонов имеет следующие особенности:

* разделение на типы может производиться по префиксу (например, `ROOT/(pages|layouts|includes)/...`) или по суффиксу (например, `ROOT/PATH/name(|.layout|.include).tmpl`)
* доступ к файлам производится через интерфейс [lookupfs.FileSystem](https://godoc.org/github.com/apisite/tpl2x/lookupfs#FileSystem) и поддерживается работа как с обычной так и со встроенной файловой системой

По имени файла шаблона формируется имя, по которому на него можно ссылаться директивой `{{template}}` и использовать в роутинге HTTP-сервера. Для этого с именем файла производятся действия:

* перевод абсолютного пути к файлу в относительный путь от корня дерева шаблонов
* удаление расширения файла
* конвертация разделителей каталогов текущей ОС в `/`
* замена всех подстрок `/__` на `/:` (':' используется в gin для обозначения параметров запроса)
* удаление начального `/` (если имя не состоит только из него)

Для случая, когда шаблон страницы не должен включаться в роутинг, используется параметр конфигурации `HidePrefix (default:".")`, такие шаблоны предназначены для прямых вызовов из кода (например, `pages/.404.tmpl`).

См. также:

* [Примеры с разделением по префиксу и суффиксу](https://godoc.org/github.com/apisite/tpl2x/lookupfs#pkg-examples)
* [Пример работы с обычной ФС](https://github.com/apisite/tpl2x/blob/master/tpl2x_test.go)
* [Пример работы с встроенной ФС](https://godoc.org/github.com/apisite/tpl2x#example-package--Execute)

### samplemeta

При формировании контента может быть создан не только он сам (т.е., фактически, тестовая строка), но и некоторые метаданные (например, заголовок страницы, имя макета, список JS-файлов для включения и т.п). Для передачи этой информации из страницы в макет используется объект структуры, соответствующей интерфейсу [tpl2x.MetaData](https://godoc.org/github.com/apisite/tpl2x#MetaData). Этот объект создается при каждом запросе страницы и интерфейс содержит только три метода:

* `SetError(error)` - вызывается внутри tpl2x при возникновении необработаннной ошибки выполнения шаблона страницы
* `Error() error` - позволяет получить эту ошибку
* `Layout()` - имя макета, который будет вызван на шаге 2

Библиотека содержит базовый пример такой структуры - [samplemeta](https://github.com/apisite/tpl2x/blob/master/samplemeta/meta.go).

В [gin-tpl2x](https://godoc.org/github.com/apisite/tpl2x/gin-tpl2x) интерфейс [gintpl2x.MetaData](https://godoc.org/github.com/apisite/tpl2x/gin-tpl2x#MetaData) дополнен функциями для формирования заголовка HTTP-ответа:

* `Status() int` - статус
* `ContentType() string` - тип контента
* `Location() string` - URL для переадресации
 
Пример реализации структуры - [gin-tpl2x/samplemeta](https://github.com/apisite/tpl2x/blob/master/gin-tpl2x/samplemeta/meta.go).

## См. также

* https://stackoverflow.com/questions/42747183/how-to-render-templates-to-multiple-layouts-in-go
* https://medium.com/@leeprovoost/dealing-with-go-template-errors-at-runtime-1b429e8b854a

##  Лицензия

The MIT License (MIT), see [LICENSE](LICENSE).

Copyright (c) 2019 Aleksei Kovrizhkin <lekovr+apisite@gmail.com>
