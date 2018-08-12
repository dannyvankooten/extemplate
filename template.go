package grender

import (
	"bufio"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var extendsRegex *regexp.Regexp

type Extemplate struct {
	set map[string]*template.Template
}

func New() *Extemplate {
	return &Extemplate{
		set: make(map[string]*template.Template),
	}
}

func (x *Extemplate) Delims(left, right string) *Extemplate {
	// TODO: Fill this func
	return x
}

func (x *Extemplate) Funcs(funcMap template.FuncMap) *Extemplate {
	// TODO: this func
	return x
}

func (x *Extemplate) Lookup(name string) *template.Template {
	// TODO this func
	return nil
}

func init() {
	var err error
	extendsRegex, err = regexp.Compile(`\{\{\/\* *?extends +?"(.+?)" *?\*\/\}\}`)
	if err != nil {
		panic(err)
	}
}

type templatefile struct {
	file    string
	name    string
	extends string
}

func ParseDir(r string) map[string]*template.Template {
	set := map[string]*template.Template{}
	sharedTemplates := template.New("")
	files := make([]*templatefile, 0)

	// find all template files
	err := filepath.Walk(r, func(path string, info os.FileInfo, err error) error {
		// skip dirs as they can never be valid templates
		if info == nil || info.IsDir() {
			return nil
		}

		// TODO make this configurable
		ext := filepath.Ext(path)
		if ext != ".html" && ext != ".tmpl" {
			return nil
		}

		layout := getLayoutForTemplate(path)
		name := strings.TrimPrefix(path, r)
		files = append(files, &templatefile{path, name, layout})
		return nil
	})

	if err != nil {
		// TODO: Handle error
	}

	// parse all templates into a single template (without inheritance)
	for _, f := range files {
		if f.extends != "" {
			continue
		}
		b, err := ioutil.ReadFile(f.file)
		if err != nil {
			// TODO: handle error
		}
		sharedTemplates.New(f.name).Parse(string(b))
	}

	// then, parse all templates again but with inheritance
	var b []byte
	for _, f := range files {
		// get template name: root/users/detail.html => users/detail.html
		tmpl := template.Must(sharedTemplates.Clone()).New(f.name)

		// TODO: allow multi-leveled extending

		// parse layout file first, because we want child template to override defined templates
		if f.extends != "" {
			layoutFile := filepath.Join(r, f.extends)
			b, _ = ioutil.ReadFile(layoutFile)
			tmpl.Parse(string(b)) // TODO: check err
		}

		// then, parse child-template
		b, _ = ioutil.ReadFile(f.file)
		tmpl.Parse(string(b))

		// add to set under normalized name (path from root)
		set[f.name] = tmpl
	}

	return set
}

// getLayoutForTemplate scans the first line of the template file for the extends keyword
func getLayoutForTemplate(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	b := scanner.Bytes()
	if l := extendsRegex.FindSubmatch(b); l != nil {
		return string(l[1])
	}

	return ""
}
