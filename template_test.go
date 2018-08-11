package grender

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseDir(t *testing.T) {
	ParseDir("examples/")
}

func TestTemplate(t *testing.T) {
	tmpl := ParseDir("examples/")

	tests := map[string]string{
		"hello.tmpl":              "Hello !",                               // normal template, no inheritance
		"child.tmpl":              "Hello world from the master template.", // template with inheritance
		"master.tmpl":             "Bye world from the master template.",   // normal template with {{ block }}
		"child-with-partial.tmpl": "Hello world! How are we today? from the master template.",
	}

	for k, v := range tests {
		if _, ok := tmpl[k]; !ok {
			t.Errorf("template not found in set: %s", k)
		}

		var buf bytes.Buffer
		if err := tmpl[k].Execute(&buf, nil); err != nil {
			t.Errorf("error executing template %s: %s", k, err)
		}

		e := strings.TrimSpace(buf.String())
		if e != v {
			t.Errorf("incorrect template result. \nExpected: %s\nActual: %s", v, e)
		}
	}

}

func BenchmarkGrenderGetLayoutForFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getLayoutForTemplate("examples/child.tmpl")
	}
}

func BenchmarkGrenderCompileTemplatesFromDir(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDir("examples/")
	}
}
