package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"

	"gitlab.com/fer-go/present/data"
	"gitlab.com/fer-go/present/types"
)

//go:embed web.tmpl
var webTemplate []byte

//go:embed web.css
var cssFile []byte

//go:embed web.js
var jsFile []byte

//go:embed web.cast.js
var jsFileCast []byte

type TemplateData struct {
	Slides []types.Slide
	Title  string
}

func init() {
	webTemplate = append(webTemplate, []byte("\n"+`{{ define "css" }}`+"\n")...)
	webTemplate = append(webTemplate, cssFile...)
	webTemplate = append(webTemplate, []byte(`{{ end }}`+"\n")...)
	webTemplate = append(webTemplate, []byte("\n"+`{{ define "js" }}`+"\n")...)
	webTemplate = append(webTemplate, jsFile...)
	webTemplate = append(webTemplate, jsFileCast...)
	webTemplate = append(webTemplate, []byte(`{{ end }}`+"\n")...)
}

func newUI() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slides := data.Presentation()
		for i := range slides {
			slides[i].PageIndex = i
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		tmpl, err := template.New("web").Parse(string(webTemplate))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// err = tmpl.Execute(w, TemplateData{
		// 	Slides: slides,
		// 	Title:  "A DEMO",
		// })
		// if err != nil {
		// 	http.Error(w, "Failed to write response", http.StatusInternalServerError)
		// 	log.Println(err)
		// }
		var out bytes.Buffer
		err = tmpl.Execute(&out, TemplateData{
			Slides: slides,
			Title:  "A DEMO",
		})
		if err != nil {
			log.Println(err)
			return
		}
		str := out.String()
		str = strings.ReplaceAll(str, " ", "")

		fmt.Print(str)
		_, err = w.Write([]byte(str))
		if err != nil {
			log.Println(err)
			return
		}

	})
}
