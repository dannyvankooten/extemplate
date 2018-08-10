package grender

import (
	"bufio"
	"fmt"
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
	set := template.New("")

	// replace existing templates.
	// NOTE: this is unsafe, but Debug should really not be true in production environments.
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

		fmt.Printf("\n")
		fmt.Printf("Processing %s\n", path)

		// get template name: root/users/detail.html => users/detail.html
		name := strings.TrimPrefix(path, r)
		layout := getLayoutForTemplate(path)
		tmpl := template.New(name)

		//tmpl := template.Must(baseTmpl.Clone()).New(name)

		templateFiles := []string{path}

		if layout != "" {
			layoutFile := filepath.Join(filepath.Dir(path), layout)
			templateFiles = append(templateFiles, layoutFile)
			// todo: keep looking for template files
		}

		fmt.Printf("Templates for %s: %#v\n", name, templateFiles)

		// parse templates in reverse order
		// this is important because we need templates defined in child files to override templates defined in parent files
		//for i, j := 0, len(templateFiles); i < j; i, j = i+1, j-1 {
		//for j := len(templateFiles) - 1; j >= 0; j-- {
		for j, _ := range templateFiles {
			fmt.Printf("Parsing %s\n", templateFiles[j])
			b, _ := ioutil.ReadFile(templateFiles[j])
			tmpl.Parse(string(b))
		}

		//tmpl = template.Must(tmpl.ParseFiles(templateFiles...))
		fmt.Printf("Template for %s: %#v\n", path, tmpl.DefinedTemplates())

		templateSet[name] = tmpl
		return nil
	})
	if err != nil {
		// TODO: Handle error
	}

	fmt.Printf("%#v", set)

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
