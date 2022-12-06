This repository moved to https://git.sr.ht/~dvko/extemplate on 2022-12-06 :warning:

---

# Extemplate [![GoDoc](http://godoc.org/github.com/dannyvankooten/extemplate?status.svg)](http://godoc.org/github.com/dannyvankooten/extemplate)  [![Build Status](https://travis-ci.org/dannyvankooten/extemplate.svg)](https://travis-ci.org/dannyvankooten/extemplate) [![Go Report Card](https://goreportcard.com/badge/github.com/dannyvankooten/extemplate)](https://goreportcard.com/report/github.com/dannyvankooten/extemplate) [![Coverage](https://gocover.io/_badge/github.com/dannyvankooten/extemplate)](https://gocover.io/github.com/dannyvankooten/extemplate)

Extemplate is a small wrapper package around [html/template](https://golang.org/pkg/html/template/) to allow for easy file-based template inheritance.

File: `templates/parent.tmpl`
```text
<html>
<head>
	<title>{{ block "title" }}Default title{{ end }}</title>
</head>
<body>
	{{ block "content" }}Default content{{ end }} 
</body>
</html>
```

File: `templates/child.tmpl`
```text
{{ extends "parent.tmpl" }}
{{ define "title" }}Child title{{ end }}
{{ define "content" }}Hello world!{{ end }}
```

File: `main.go`
```go
xt := extemplate.New()
xt.ParseDir("templates/", []string{".tmpl"})
_ = xt.ExecuteTemplate(os.Stdout, "child.tmpl", "no data needed") 
// Output: <html>.... Hello world! ....</html>
```

Extemplate recursively walks all files in the given directory and will parse the files matching the given extensions as a template. Templates are named by path and filename, relative to the root directory.

For example, calling `ParseDir("templates/", []string{".tmpl"})` on the following directory structure:

```text
templates/
  |__ admin/
  |      |__ index.tmpl
  |      |__ edit.tmpl
  |__ index.tmpl
```

Will result in the following templates:

```text
admin/index.tmpl
admin/edit.tmpl
index.tmpl
```

Check out the [tests](https://github.com/dannyvankooten/extemplate/blob/master/template_test.go) and [examples directory](https://github.com/dannyvankooten/extemplate/tree/master/examples) for more examples.

### Benchmarks

You will most likely never have to worry about performance, when using this package properly. 
The benchmarks are purely listed here so we have a place to keep track of progress.

```
BenchmarkExtemplateGetLayoutForTemplate-8   	 2000000	       923 ns/op	     104 B/op	       3 allocs/op
BenchmarkExtemplateParseDir-8               	    5000	    227898 ns/op	   34864 B/op	     325 allocs/op
```

### License

MIT
