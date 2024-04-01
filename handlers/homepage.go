package handlers

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"text/template"

	"gitlab.com/fer-go/present/data"
	"gitlab.com/fer-go/present/types"
	"gitlab.com/fer-go/present/ui"
)

type TemplateData struct {
	Slides []types.Slide
	Menu   []types.Menu
	Title  string
	Port   int
}

func Homepage(port int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		presentation := data.Presentation()
		slides := presentation.Slides
		for i := range slides {
			slides[i].PageIndex = i
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		tmpl, err := template.New("web").Parse(string(ui.WebTemplate()))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var out bytes.Buffer
		err = tmpl.Execute(&out, TemplateData{
			Slides: slides,
			Title:  presentation.Title,
			Menu:   presentation.Menu,
			Port:   port,
		})
		if err != nil {
			log.Println(err)
			return
		}
		str := out.String()
		str = strings.ReplaceAll(str, " ", "")

		// fmt.Print(str)
		_, err = w.Write([]byte(str))
		if err != nil {
			log.Println(err)
			return
		}
	})
}
