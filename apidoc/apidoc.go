package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"

	"go/doc"
	"go/parser"
	"go/token"
)

var (
	marker      = flag.String("marker", ":WEBAPI:", "Marker in docstring")
	moduleDir   = flag.String("dir", "", "Path")
	templateDir = flag.String("template-dir", "", "Template directory")
)

type Param struct {
	Type        string
	Name        string
	Description string
	Default     string
}

type QueryParams []Param
func (s QueryParams) Less(i, j int) bool { return s[i].Name < s[j].Name }

type WebAPI struct {
	URL         string
	Method      string
	Description string
	Arguments   QueryParams
	Parameters  QueryParams
}

type WAPI []WebAPI
func (s WAPI) Len() int      { return len(s) }
func (s WAPI) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByMethod struct{ WAPI }
func (s ByMethod) Less(i, j int) bool { return s.WAPI[i].Method < s.WAPI[j].Method }

func fileExists(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return os.IsExist(err)
	}
	file.Close()
	return true
}

func main() {
	flag.Parse()

	fset := token.NewFileSet()
	f, err := parser.ParseDir(fset, *moduleDir, nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	groupByURL := make(map[string][]WebAPI)

	for k, pkg := range f {
		p := doc.New(pkg, k, 0)

		for _, f := range p.Funcs {
			fDoc := f.Doc[:]
			for {
				idx := strings.Index(fDoc, *marker)
				if idx < 0 {
					break
				}
				fDoc = fDoc[idx+len(*marker):]
				i := idx+len(*marker)+1

				for i < len(fDoc) {
					var jsonDoc WebAPI
					if err := json.Unmarshal([]byte(fDoc[:i]), &jsonDoc); err == nil {
						groupByURL[jsonDoc.URL] = append(groupByURL[jsonDoc.URL], jsonDoc)
						break
					}
					i += 1
				}
				if i == len(fDoc) {
					break
				}
				fDoc = fDoc[i:]
			}
		}
	}

	apiTmpl, err := template.ParseFiles(*templateDir + "/webapi-call.html")
	if err != nil {
		panic(err)
	}

	if fileExists(*templateDir + "/page-start.html") {
		tmpl, err := template.ParseFiles(*templateDir + "/page-start.html")
		if err != nil {
			panic(err)
		}
		if err := tmpl.Execute(os.Stdout, nil); err != nil {
			panic(err)
		}
	}

	var urls sort.StringSlice
	for url := range groupByURL {
		urls = append(urls, url)
	}
	urls.Sort()

	for _, url := range urls {
		group := groupByURL[url]

		if fileExists(*templateDir + "/group-start.html") {
			tmpl, err := template.ParseFiles(*templateDir + "/group-start.html")
			if err != nil {
				panic(err)
			}
			if err := tmpl.Execute(os.Stdout, url); err != nil {
				panic(err)
			}
		}

		sort.Sort(ByMethod{group})
		for _, api := range group {
			if err := apiTmpl.Execute(os.Stdout, api); err != nil {
				panic(err)
			}
		}

		if fileExists(*templateDir + "/group-end.html") {
			tmpl, err := template.ParseFiles(*templateDir + "/group-end.html")
			if err != nil {
				panic(err)
			}
			if err := tmpl.Execute(os.Stdout, url); err != nil {
				panic(err)
			}
		}
	}

	if fileExists(*templateDir + "/page-end.html") {
		tmpl, err := template.ParseFiles(*templateDir + "/page-end.html")
		if err != nil {
			panic(err)
		}
		if err := tmpl.Execute(os.Stdout, nil); err != nil {
			panic(err)
		}
	}
}
