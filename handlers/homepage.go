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

func Homepage(port int, loginPage []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// user, pass, _ := r.BasicAuth()
		// var cookie *http.Cookie
		// var err error
		// if user == "" || pass == "" {
		// 	cookie, err = r.Cookie("present")
		// 	if err == nil {
		// 		// Cookie exists, you can access its value using cookie.Value
		// 		fmt.Println("Cookie value:", cookie.Value)
		// 		user = "present"
		// 		pass = cookie.Value
		// 	}
		// }
		// if user != "present" || pass != os.Getenv("USER_PWD") {
		// 	_, err := w.Write(loginPage)
		// 	if err != nil {
		// 		log.Println(err)
		// 		return
		// 	}
		// 	return
		// }
		// cookieSet := http.Cookie{
		// 	Name:  "present",
		// 	Value: os.Getenv("USER_PWD"),
		// }
		// http.SetCookie(w, &cookieSet)
		// if cookie == nil {
		// 	return
		// }

		presentation := data.Presentation()
		slides := presentation.Slides
		for i := range slides {
			slides[i].PageIndex = i
		}
		shiftPage := 0
		for i := 1; i < len(presentation.Slides)-1; i++ {
			if presentation.Slides[i].PrintOnly {
				shiftPage++
			}
			presentation.Slides[i].PageIndex -= shiftPage
		}

		shiftPage = 0
		for i := 1; i < len(presentation.Slides)-1; i++ {
			presentation.Slides[i].PageNumber -= shiftPage
			if presentation.Slides[i].PrintDisable {
				shiftPage++
			}
		}
		for i := 1; i < len(presentation.Slides)-1; i++ {
			presentation.Slides[i].PageID = i
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
