package grender

import (
	"bytes"
	"fmt"
	"testing"
)

func TestParseDir(t *testing.T) {
	ParseDir("examples/")
}

func TestTemplate(t *testing.T) {
	tmpl := ParseDir("examples/")

	if _, ok := tmpl["child.tmpl"]; !ok {
		t.Errorf("template not found")
	} else {
		var buf bytes.Buffer
		if err := tmpl["child.tmpl"].Execute(&buf, nil); err != nil {
			t.Errorf("error executing child.tmpl: %s", err)
		}
		fmt.Printf("Executed template %s: %s\n", "child.tmpl", string(buf.Bytes()))
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
