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

func init() {
	var err error
	extendsRegex, err = regexp.Compile(`\{\{\/\* *?extends +?"(.+?)" *?\*\/\}\}`)
	if err != nil {
		panic(err)
	}
}

func ParseDir(r string) map[string]*template.Template {
	templateSet := map[string]*template.Template{}
	all := template.New("")
	files := []string{}

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

		files = append(files, path)
		return nil
	})

	if err != nil {
		// TODO: Handle error
	}

	// parse all templates into a single template (without inheritance)
	for _, f := range files {
		b, err := ioutil.ReadFile(f)
		if err != nil {
			// TODO: handle error
		}
		all.Parse(string(b))
	}

	// then, parse all templates again but with inheritance
	for _, path := range files {
		// get template name: root/users/detail.html => users/detail.html
		name := strings.TrimPrefix(path, r)
		layout := getLayoutForTemplate(path)
		//tmpl := template.New(name)
		tmpl := template.Must(all.Clone()).New(name)
		//tmpl := template.Must(baseTmpl.Clone()).New(name)

		templateFiles := []string{path}

		if layout != "" {
			layoutFile := filepath.Join(filepath.Dir(path), layout)
			templateFiles = append(templateFiles, layoutFile)
			// todo: keep looking for layout files
		}

		// parse templates in reverse order
		// this is important because we need templates defined in child files to override templates defined in parent files
		for j := len(templateFiles) - 1; j >= 0; j-- {
			//for j, _ := range templateFiles {
			b, _ := ioutil.ReadFile(templateFiles[j])
			tmpl.Parse(string(b))
		}

		templateSet[name] = tmpl
	}

	return templateSet
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
