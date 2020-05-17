package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var flagPackage = flag.String("p", "main", "package of static assets")
var flagFilename = flag.String("o", "static_templates.go", "output filename")
var flagExtensions = flag.String("e", "html,tmpl", "included file extensions")

func customUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] directory\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = customUsage
	flag.Parse()

	if err, showUsage := bundleTemplates(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		if showUsage {
			flag.Usage()
		}
		os.Exit(1)
	}
}

func bundleTemplates() (error, bool) {
	if flag.NArg() == 0 {
		return errors.New("a root directory must be provided"), true
	}

	exts := strings.Split(*flagExtensions, ",")

	root := flag.Arg(0)
	f, err := scanFiles(root, exts)
	if err != nil {
		return err, false
	}

	if err := writeStatic(*flagFilename, f); err != nil {
		return err, false
	}

	return nil, false
}

func scanFiles(root string, extensions []string) (string, error) {
	var files = make(map[string][]byte)
	var exts = map[string]bool{}

	// ensure root has trailing slash
	root = strings.TrimSuffix(root, "/") + "/"

	// create map of allowed extensions
	for _, e := range extensions {
		exts["."+e] = true
	}

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

		// read file into memory
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		name := strings.TrimPrefix(path, root)

		files[name] = contents

		return nil
	})

	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", errors.New("no template files found")
	}

	b, err := json.Marshal(files)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeStatic(filename, staticTemplate string) error {
	tpl := `package %s

import "github.com/dannyvankooten/extemplate"

func init() {
	extemplate.StaticTemplates(%s)
}
`
	output := fmt.Sprintf(tpl, *flagPackage, "`"+staticTemplate+"`")

	return ioutil.WriteFile(filename, []byte(output), 0644)
}
