package extemplate

import (
	"bytes"
	"embed"
	"html/template"
	"strings"
	"sync"
	"testing"
)

var x *Extemplate
var xFS *Extemplate
var once sync.Once

//go:embed examples
var templateFS embed.FS

func newXTmpl() *Extemplate {
	return New().Delims("{{", "}}").Funcs(template.FuncMap{
		"tolower": strings.ToLower,
	})
}

func setup() {
	x = newXTmpl()

	err := x.ParseDir("examples", []string{".tmpl"})
	if err != nil {
		panic(err)
	}

	xFS = newXTmpl()
	err = x.ParseFS(templateFS, []string{".tmpl"})
	if err != nil {
		panic(err)
	}
}

func TestLookup(t *testing.T) {
	once.Do(setup)

	if tmpl := x.Lookup("foobar"); tmpl != nil {
		t.Errorf("Lookup: expected nil, got %#v", tmpl)
	}

	if tmpl := x.Lookup("child.tmpl"); tmpl == nil {
		t.Error("Lookup: expected template, got nil")
	}
}

func TestExecuteTemplate(t *testing.T) {
	once.Do(setup)

	var buf bytes.Buffer
	if err := x.ExecuteTemplate(&buf, "child.tmpl", nil); err != nil {
		t.Errorf("ExecuteTemplate: %s", err)
	}
	if err := x.ExecuteTemplate(&buf, "foobar", nil); err == nil {
		t.Error("ExecuteTemplate: expected err for unexisting template, got none")
	}

}

func TestTemplates(t *testing.T) {
	once.Do(setup)

	tests := map[string]string{
		"parent.tmpl":      "Hello from master.tmpl",                                     // normal template with {{ block }}
		"child.tmpl":       "Hello from child.tmpl\n\tHello from partials/question.tmpl", // template with inheritance
		"grand-child.tmpl": "Hello from grand-child.tmpl",                                // template with nested inheritance
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
		e = strings.Replace(e, "\r\n", "\n", -1)
		if e != v {
			t.Errorf("incorrect template result. \nExpected: %s\nActual: %s", v, e)
		}
	}
}

func TestNewTemplateFile(t *testing.T) {
	tests := map[string]string{
		"{{ extends \"foo.html\" }}": "foo.html",
		"Nothing":                    "",
		"{{ extends \"dir/file.html\" }}\n {{ .Var }}": "dir/file.html",
	}

	for c, e := range tests {
		tf, err := newTemplateFile([]byte(c))
		if err != nil {
			t.Error(err)
		}
		if tf.layout != e {
			t.Errorf("Expected layout %s, got %s", e, tf.layout)
		}
	}
}

func BenchmarkExtemplateGetLayoutForTemplate(b *testing.B) {
	c := []byte("{{ extends \"foo.html\" }}")
	for i := 0; i < b.N; i++ {
		if _, err := newTemplateFile(c); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkExtemplateParseDir(b *testing.B) {
	x := New().Funcs(template.FuncMap{
		"foo": strings.ToLower,
	})
	for i := 0; i < b.N; i++ {
		x.ParseDir("examples", []string{".tmpl"})
	}
}

func BenchmarkExtemplateParseFS(b *testing.B) {
	x := New().Funcs(template.FuncMap{
		"foo": strings.ToLower,
	})

	for i := 0; i < b.N; i++ {
		x.ParseFS(templateFS, []string{".tmpl"})
	}
}
