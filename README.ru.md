# tpl2x
> golang-библиотека для трансляции дерева шаблонов в адреса страниц сайта.

[![GoDoc][gd1]][gd2]
 [![codecov][cc1]][cc2]
 [![GoCard][gc1]][gc2]
 [![GitHub Release][gr1]][gr2]
 ![GitHub code size in bytes][sz]
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

* Статус проекта: код готов к релизу
* Текущая версия: 0.2.1
* Планы до релиза (v 1.0):
  * [ ] актуализировать документацию
  * [ ] определиться с именем репозитория (apitpl? tpltree?)
  * [ ] решить вопрос хранения имени метода `content` 

Библиотека предназначена для построения вебсайтов, у которых дерево страниц (или его часть) совпадает с деревом файлов (шаблонов страниц) в некотором каталоге. Первичным таким проектом является [apisite](https://github.com/apisite/apisite), где шаблоны получают данные с помощью вызовов API и страницы может формировать универсальный обработчик.

## Структура

Библиотека разделена на следующие части:

* [tpl2x](https://godoc.org/github.com/apisite/tpl2x) - код формирования страницы из шаблона (выполняется в два этапа - формирование контента и сборка страницы по макеты с включением в него контента)
* [lookupfs](https://godoc.org/github.com/apisite/tpl2x/lookupfs) - получение из файловой системы (обычной или встроенной) списков шаблонов
* [samplemeta](https://godoc.org/github.com/apisite/tpl2x/samplemeta) - пример метаданных, которые могут передаваться из шаблона контента в шаблон макета
* [samplefs](https://godoc.org/github.com/apisite/tpl2x/samplefs) - встроенная файловая система для сокращения кода примеров, используется только в `example_*_test.go` 
* [gin-tpl2x](https://godoc.org/github.com/apisite/tpl2x/gin-tpl2x) - интеграция функционала tpl2x в [gin](https://github.com/gin-gonic/gin) (код оформлен модулем, чтобы его зависимости не попали в остальные части), для тестов и примеров этот модуль имеет свои копии [samplemeta](https://godoc.org/github.com/apisite/tpl2x/gin-tpl2x/samplemeta) и [samplefs](https://godoc.org/github.com/apisite/tpl2x/gin-tpl2x/samplefs)

## Особенности реализации

### Шаблоны страниц

Дерево шаблонов (т.е. и каталогов с файлами и страниц сайта) разбито на три группы:

* страницы (pages) - шаблоны, формирующие контент страницы (по пути к файлу формируется URL страницы)
* макеты (layouts) - шаблоны, состоящие из общих для всех страниц элементов (шапки, подвала и т.п.), в которые помещается контент (кроме этой вставки, шаблон может быть обычной HTML-страницей)
* включения (includes) - блоки, которые могут использоваться в страницах и макетах

См. также: [Пример дерева шаблонов](https://github.com/apisite/tpl2x/tree/master/gin-tpl2x/testdata)

#### Обновление шаблонов

Обновление шаблонов производится без перекомпиляция кода (кроме случая, когда используется встроенная ФС). А если установлен флаг `ParseAlways(true)`, то для обновления (кроме изменения списка страниц, когда его надо регистрировать в роутере) не требуется и рестарт приложения.

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

Получение списка файлов для каждой из трех групп шаблонов имеет следующие особенности:

* разделение на группы может производиться по префиксу (например, `ROOT/(pages|layouts|includes)/...`) или по суффиксу (например, `ROOT/PATH/name(|.layout|.include).tmpl`)
* доступ к файлам производится через интерфейс [lookupfs.FileSystem](https://godoc.org/github.com/apisite/tpl2x/lookupfs#FileSystem) и поддерживается работа как с обычной так и со встроенной файловой системой

По имени файла шаблона формируется имя, по которому на него можно ссылаться директивой `{{template}}` и использовать в роутинге HTTP-сервера. Для этого с именем файла производятся действия:

* перевод абсолютного пути к файлу в относительный путь от корня дерева шаблонов
* удаление расширения файла
* конвертация разделителей каталогов текущей ОС в `/`
* замена всех подстрок `/__` на `/:` (':' используется в gin для обозначения параметров запроса)
* удаление начального `/` (если имя не состоит только из него)

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

Copyright (c) 2018 Aleksei Kovrizhkin <lekovr+apisite@gmail.com>
