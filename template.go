// Copyright 2017 Danny van Kooten. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package extemplate

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var extendsRegex *regexp.Regexp

// Extemplate holds a reference to all templates
// and shared configuration like Delims or FuncMap
type Extemplate struct {
	shared    *template.Template
	templates map[string]*template.Template
}

type templatefile struct {
	file    string
	name    string
	extends string
}

func init() {
	var err error
	extendsRegex, err = regexp.Compile(`\{\{\/\* *?extends +?"(.+?)" *?\*\/\}\}`)
	if err != nil {
		panic(err)
	}
}

// New allocates a new, empty, template map
func New() *Extemplate {
	return &Extemplate{
		shared:    template.New(""),
		templates: make(map[string]*template.Template),
	}
}

// Delims sets the action delimiters to the specified strings,
// to be used in subsequent calls to ParseDir.
// Nested template  definitions will inherit the settings.
// An empty delimiter stands for the corresponding default: {{ or }}.
// The return value is the template, so calls can be chained.
func (x *Extemplate) Delims(left, right string) *Extemplate {
	x.shared.Delims(left, right)
	return x
}

// Funcs adds the elements of the argument map to the template's function map.
// It must be called before templates are parsed
// It panics if a value in the map is not a function with appropriate return
// type or if the name cannot be used syntactically as a function in a template.
// It is legal to overwrite elements of the map. The return value is the Extemplate instance,
// so calls can be chained.
func (x *Extemplate) Funcs(funcMap template.FuncMap) *Extemplate {
	x.shared.Funcs(funcMap)
	return x
}

// Lookup returns the template with the given name
// It returns nil if there is no such template or the template has no definition.
func (x *Extemplate) Lookup(name string) *template.Template {
	if t, ok := x.templates[name]; ok {
		return t
	}

	return nil
}

// ExecuteTemplate applies the template named name to the specified data object and writes the output to wr.
func (x *Extemplate) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	tmpl := x.Lookup(name)
	if tmpl == nil {
		return fmt.Errorf("extemplate: no template %q", name)
	}

	return tmpl.Execute(wr, data)
}

// ParseDir walks the given directory root and parses all files with any of the registered extensions.
// Default extensions are .html and .tmpl
// If a template file has {{/* extends "other-file.tmpl" */}} as its first line it will parse that file for base templates.
// Parsed templates are named relative to the given root directory
func (x *Extemplate) ParseDir(root string, extensions []string) error {
	var b []byte
	var err error

	files, err := findTemplateFiles(root, extensions)
	if err != nil {
		return err
	}

	// parse all non-child templates into the shared template namespace
	for _, f := range files {
		if f.extends != "" {
			continue
		}
		b, err = ioutil.ReadFile(f.file)
		if err != nil {
			return err
		}

		_, err = x.shared.New(f.name).Parse(string(b))
		if err != nil {
			return err
		}
	}

	// then, parse all templates again but with inheritance
	for _, f := range files {
		// get template name: root/users/detail.html => users/detail.html
		tmpl := template.Must(x.shared.Clone()).New(f.name)

		// TODO: allow multi-leveled extending

		// parse layout file first, because we want child template to override defined templates
		if f.extends != "" {
			layoutFile := filepath.Join(root, f.extends)
			b, err = ioutil.ReadFile(layoutFile)
			if err != nil {
				return err
			}
			_, err = tmpl.Parse(string(b))
			if err != nil {
				return err
			}
		}

		// then, parse child-template
		b, _ = ioutil.ReadFile(f.file)
		_, err = tmpl.Parse(string(b))
		if err != nil {
			return err
		}

		// add to set under normalized name (path from root)
		x.templates[f.name] = tmpl
	}

	return nil
}

func findTemplateFiles(root string, extensions []string) ([]*templatefile, error) {
	var files = make([]*templatefile, 0)
	var exts = map[string]bool{}

	// create map of allowed extensions
	for _, e := range extensions {
		exts[e] = true
	}

	// find all template files
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// skip dirs as they can never be valid templates
		if info == nil || info.IsDir() {
			return nil
		}

		// skip if extension not in list of allowed extensions
		e := filepath.Ext(path)
		if _, ok := exts[e]; !ok {
			return nil
		}

		layout, err := getLayoutForTemplate(path)
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(path, root)
		files = append(files, &templatefile{path, name, layout})
		return nil
	})

	return files, err
}

// getLayoutForTemplate scans the first line of the template file for the extends keyword
func getLayoutForTemplate(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	b := scanner.Bytes()
	if l := extendsRegex.FindSubmatch(b); l != nil {
		return string(l[1]), nil
	}

	return "", nil
}
