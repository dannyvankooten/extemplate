# Extemplate [![GoDoc](http://godoc.org/github.com/dannyvankooten/extemplate?status.svg)](http://godoc.org/github.com/dannyvankooten/extemplate)  [![Build Status](https://travis-ci.org/dannyvankooten/extemplate.svg)](https://travis-ci.org/dannyvankooten/extemplate) [![Go Report Card](https://goreportcard.com/badge/github.com/dannyvankooten/extemplate)](https://goreportcard.com/report/github.com/dannyvankooten/extemplate)

Extemplate is a small wrapper package around [html/template](https://golang.org/pkg/html/template/) to allow for easy file-based template inheritance.

File: `templates/parent.tmpl`
```text
{{block "content"}}Bye{{end}} world
```

File: `templates/child.tmpl`
```text
{{/* extends "parent.tmpl" */}}
{{define "content"}}Hello{{end}}
```

```go
xt := extemplate.New()
xt.ParseDir("templates/", []string{".tmpl"})
_ = xt.ExecuteTemplate(os.Stdout, "child.tmpl", "no data needed")
```

Check out the [tests](https://github.com/dannyvankooten/extemplate/blob/master/template_test.go) and [examples directory](https://github.com/dannyvankooten/extemplate/tree/master/examples) for more examples.

### License

MIT
