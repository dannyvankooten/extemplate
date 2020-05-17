// Copyright 2017 Danny van Kooten. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package extemplate

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var staticTemplates string
var extendsRegex *regexp.Regexp

// Extemplate holds a reference to all templates
// and shared configuration like Delims or FuncMap
type Extemplate struct {
	shared       *template.Template
	sharedBackup *template.Template
	templates    map[string]*template.Template
	autoReload   bool
	root         string
	extensions   []string
	fileTreeHash []byte
}

type templatefile struct {
	contents []byte
	layout   string
}

func init() {
	var err error
	extendsRegex, err = regexp.Compile(`\{\{ *?extends +?"(.+?)" *?\}\}`)
	if err != nil {
		panic(err)
	}
}

// New allocates a new, empty, template map
func New() *Extemplate {
	shared := template.New("")
	return &Extemplate{
		shared:    shared,
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

// AutoReload configures whether the templates will be automatically reloaded
// from disk when they change. This setting has no effect when static templates
// are compiled into the source.
func (x *Extemplate) AutoReload(reload bool) {
	x.autoReload = reload
}

// ExecuteTemplate applies the template named name to the specified data object and writes the output to wr.
// Templates will be automatically reloaded if AutoReload is true.
// In addition to handling a single data object as is done in the standard library, this
// wrapper also supports providing key/value pairs as individual arguments. e.g.
//
// tpl.ExecuteTemplate(w, "template.html", "name", "Bob", "age", 42)
func (x *Extemplate) ExecuteTemplate(wr io.Writer, name string, data ...interface{}) error {
	if staticTemplates == "" && x.autoReload {
		if err := x.reloadTemplates(); err != nil {
			return err
		}
	}

	tmpl := x.Lookup(name)
	if tmpl == nil {
		return fmt.Errorf("extemplate: no template %q", name)
	}

	switch len(data) {
	case 0:
		return tmpl.Execute(wr, nil)
	case 1:
		return tmpl.Execute(wr, data[0])
	}

	if len(data)%2 != 0 {
		return errors.New("odd number of key/value pairs as template data")
	}

	dataMap := make(map[string]interface{})

	for i := 0; i < len(data); i += 2 {
		key, ok := data[i].(string)
		if !ok {
			return errors.New("template data key must be a string")
		}
		dataMap[key] = data[i+1]
	}

	return tmpl.Execute(wr, dataMap)
}

// ParseDir walks the given directory root and parses all files with any of the registered extensions.
// Default extensions are .html and .tmpl
// If a template file has {{/* extends "other-file.tmpl" */}} as its first line it will parse that file for base templates.
// Parsed templates are named relative to the given root directory
func (x *Extemplate) ParseDir(root string, extensions []string) error {
	files := make(map[string]*templatefile)

	x.root = root

	if len(extensions) == 0 {
		extensions = []string{".html", ".tmpl"}
	}
	x.extensions = extensions

	if staticTemplates != "" {
		m := make(map[string][]byte)

		err := json.Unmarshal([]byte(staticTemplates), &m)
		if err != nil {
			return err
		}
		for name, contents := range m {
			tpl, err := newTemplateFile(contents)
			if err != nil {
				return err
			}
			files[name] = tpl
		}
	} else {
		fileList, hash, err := buildFileList(root, extensions)
		if err != nil {
			return err
		}

		files, err = loadTemplateFiles(root, fileList)
		if err != nil {
			return err
		}
		x.fileTreeHash = hash
	}

	return x.parseTemplates(files)
}

func (x *Extemplate) ParseStatic(templates string) error {
	var files map[string]*templatefile

	if err := json.Unmarshal([]byte(templates), &files); err != nil {
		return err
	}
	return x.parseTemplates(files)
}

func (x *Extemplate) parseTemplates(files map[string]*templatefile) error {
	var b []byte
	var err error

	if x.sharedBackup == nil {
		if x.sharedBackup, err = x.shared.Clone(); err != nil {
			return err
		}
	}

	// parse all non-child templates into the shared template namespace
	for name, tf := range files {
		if tf.layout != "" {
			continue
		}

		_, err = x.shared.New(name).Parse(string(tf.contents))
		if err != nil {
			return err
		}
	}

	// then, parse all templates again but with inheritance
	for name, tf := range files {

		// if this is a non-child template, no need to re-parse
		if tf.layout == "" {
			x.templates[name] = x.shared.Lookup(name)
			continue
		}

		tmpl := template.Must(x.shared.Clone()).New(name)

		// add to set under normalized name (path from root)
		x.templates[name] = tmpl

		// parse parent templates
		templateFiles := []string{name}
		pname := tf.layout
		parent, parentExists := files[pname]
		for parentExists {
			templateFiles = append(templateFiles, pname)
			pname = parent.layout
			parent, parentExists = files[parent.layout]
		}

		// parse template files in reverse order (because children should override parents)
		for j := len(templateFiles) - 1; j >= 0; j-- {
			b = files[templateFiles[j]].contents
			_, err = tmpl.Parse(string(b))
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func buildFileList(root string, extensions []string) ([]string, []byte, error) {
	var files []string
	var exts = map[string]bool{}

	// ensure root has trailing slash
	root = strings.TrimSuffix(root, "/") + "/"

	// create map of allowed extensions
	for _, e := range extensions {
		exts[e] = true
	}

	hash := sha256.New()

	// find all template files
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// skip dirs as they can never be valid templates
		if info == nil || info.IsDir() {
			return nil
		}

		// skip if extension not in list of allowed extensions
		if !exts[filepath.Ext(path)] {
			return nil
		}

		// incorporate path and modification time into hash for
		// detecting updates for auto reload
		hash.Write([]byte(path))
		hash.Write([]byte(info.ModTime().String()))

		files = append(files, path)

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return files, hash.Sum(nil), nil
}

func loadTemplateFiles(root string, paths []string) (map[string]*templatefile, error) {
	// load all template files
	var files = map[string]*templatefile{}

	for _, path := range paths {
		name := strings.TrimPrefix(path, root)

		// read file into memory
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		tf, err := newTemplateFile(contents)
		if err != nil {
			return nil, err
		}

		files[name] = tf
	}

	return files, nil
}

func (x *Extemplate) reloadTemplates() error {
	fileList, hash, err := buildFileList(x.root, x.extensions)
	if err != nil {
		return err
	} else if bytes.Equal(hash, x.fileTreeHash) {
		return nil
	}

	files, err := loadTemplateFiles(x.root, fileList)
	if err != nil {
		return err
	}

	x.fileTreeHash = hash

	x.shared, err = x.sharedBackup.Clone()
	if err != nil {
		return err
	}

	x.templates = make(map[string]*template.Template)

	if err := x.parseTemplates(files); err != nil {
		return err
	}

	return nil
}

// newTemplateFile parses the file contents into something that text/template can understand
func newTemplateFile(c []byte) (*templatefile, error) {
	tf := &templatefile{
		contents: c,
	}

	r := bytes.NewReader(tf.contents)
	pos := 0
	var line []byte
	for {
		ch, l, err := r.ReadRune()
		pos += l

		// read until first line or EOF
		if ch == '\n' || err == io.EOF {
			line = c[0:pos]
			break
		}
	}

	if len(line) < 10 {
		return tf, nil
	}

	// if we have a match, strip first line of content
	if m := extendsRegex.FindSubmatch(line); m != nil {
		tf.layout = string(m[1])
		tf.contents = c[len(line):]
	}

	return tf, nil
}

func StaticTemplates(s string) {
	staticTemplates = s
}
