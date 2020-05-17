#!/bin/bash

mkdir -p static_tmp

# Statically bundle examples templates
go run cmd/exstatic/exstatic.go -o static_tmp/static_templates.go examples
cd static_tmp

# Add a go.mod file so we can test the local version of the package
cat > go.mod <<-EOF
module static

go 1.13

replace github.com/dannyvankooten/extemplate => ../

require github.com/dannyvankooten/extemplate v0.0.0-20180818082729-efbdf6eacd7e
EOF

cat > static.go <<- EOF
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/dannyvankooten/extemplate"
)

var expected = "Hello from child.tmpl\n\tHello from partials/question.tmpl"

func main() {
	var out bytes.Buffer

	xt := extemplate.New().Funcs(template.FuncMap{
		"tolower": strings.ToLower,
	})
	xt.ParseDir("this doesn't matter since they're bundled", []string{".tmpl"})
	xt.ExecuteTemplate(&out, "child.tmpl", nil)
	result := strings.TrimSpace(out.String())

	if result != expected {
		log.Fatalf("expected: %q, got %q", expected, result)
	}

	fmt.Println("PASS")
}
EOF
go run .
cd ..
rm -rf static_tmp
