package extemplate

import (
	"bytes"
	"html/template"
	"strings"
	"sync"
	"testing"
)

var x *Extemplate
var once sync.Once

func setup() {
	x = New().Delims("{{", "}}").Funcs(template.FuncMap{"foo": func() string { return "bar" }})
	err := x.ParseDir("examples/", []string{".tmpl"})
	if err != nil {
		panic(err)
	}
}

func TestLookup(t *testing.T) {
	once.Do(setup)

	if tmpl := x.Lookup("foobar"); tmpl != nil {
		t.Errorf("Lookup: expected nil, got %#v", tmpl)
	}

	if tmpl := x.Lookup("hello.tmpl"); tmpl == nil {
		t.Error("Lookup: expected template, got nil")
	}
}

func TestExecuteTemplate(t *testing.T) {
	once.Do(setup)

	var buf bytes.Buffer
	if err := x.ExecuteTemplate(&buf, "hello.tmpl", nil); err != nil {
		t.Errorf("ExecuteTemplate: %s", err)
	}
	if err := x.ExecuteTemplate(&buf, "foobar", nil); err == nil {
		t.Error("ExecuteTemplate: expected err for unexisting template, got none")
	}

}

func TestTemplates(t *testing.T) {
	once.Do(setup)

	tests := map[string]string{
		"hello.tmpl":                        "Hello from hello.tmpl",        // normal template, no inheritance
		"subdir/hello.tmpl":                 "Hello from subdir/hello.tmpl", // normal template, no inheritance
		"child.tmpl":                        "Hello from child.tmpl",        // template with inheritance
		"grand-child.tmpl":                  "Hello from grand-child.tmpl",  // template with inheritance
		"master.tmpl":                       "Hello from master.tmpl",       // normal template with {{ block }}
		"child-with-shared-components.tmpl": "Hello bar from child-with-shared-components.tmpl\n\tHello from partials/question.tmpl",
	}

	for k, v := range tests {
		tmpl := x.Lookup(k)
		if tmpl == nil {
			t.Errorf("template not found in set: %s", k)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, nil); err != nil {
			t.Errorf("error executing template %s: %s", k, err)
		}

		e := strings.TrimSpace(buf.String())
		if e != v {
			t.Errorf("incorrect template result. \nExpected: %s\nActual: %s", v, e)
		}
	}

}

func BenchmarkExtemplateGetLayoutForTemplate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getLayoutForTemplate("examples/child.tmpl")
	}
}

func BenchmarkExtemplateParseDir(b *testing.B) {
	x := New().Funcs(template.FuncMap{
		"foo": strings.ToLower,
	})
	for i := 0; i < b.N; i++ {
		x.ParseDir("examples/", []string{".tmpl"})
	}
}
