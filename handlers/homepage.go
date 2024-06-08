package handlers

import (
	"bytes"
	"cmp"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"gitlab.com/fer-go/present/data"
	"gitlab.com/fer-go/present/hash"
	"gitlab.com/fer-go/present/types"
	"gitlab.com/fer-go/present/ui"
)

type TemplateData struct {
	Slides        []types.Slide
	Menu          []types.Menu
	Title         string
	Port          int
	PageNext      []string
	PagePrevious  []string
	TerminalCast  []string
	TerminalClose []string
	MenuKey       []string
}

func Homepage(port int, loginPage []byte, userPwd, adminPwd string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		var cookie *http.Cookie
		var cookiePassword bool
		var err error
		if user == "" || pass == "" {
			cookie, err = r.Cookie("present")
			if err == nil {
				// Cookie exists, you can access its value using cookie.Value
				fmt.Println("Cookie value:", cookie.Value)
				user = "present"
				pass = cookie.Value
				cookiePassword = true
			}
		}
		pass, _ = hash.Hash(pass)
		passwordOK := hash.Equal(pass, userPwd) || hash.Equal(pass, adminPwd)
		log.Println(passwordOK)
		if passwordOK {
			_, err := w.Write(loginPage)
			if err != nil {
				log.Println(err)
				return
			}
			return
		}
		if !cookiePassword {
			pass, err = hash.Hash(pass)
			if err != nil {
				log.Println(err)
				return
			}
		}

		cookieSet := http.Cookie{
			Name:  "present",
			Value: pass,
		}
		http.SetCookie(w, &cookieSet)
		if cookie == nil {
			return
		}

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
		pageNextStr := cmp.Or(os.Getenv("NEXT_PAGE"), "ArrowRight,ArrowDown,PageDown,Space")
		pageNextStr = strings.ReplaceAll(pageNextStr, "Space", " ")
		pagePreviousStr := cmp.Or(os.Getenv("PREVIOUS_PAGE"), "ArrowLeft,ArrowUp,PageUp")
		pagePreviousStr = strings.ReplaceAll(pagePreviousStr, "Space", " ")
		terminalCastStr := cmp.Or(os.Getenv("TERMINAL_CAST"), "r")
		terminalCastStr = strings.ReplaceAll(terminalCastStr, "Space", " ")
		terminalCloseStr := cmp.Or(os.Getenv("TERMINAL_CLOSE"), "c")
		terminalCloseStr = strings.ReplaceAll(terminalCloseStr, "Space", " ")
		menuKeyStr := cmp.Or(os.Getenv("MENU_KEY"), "m")
		menuKeyStr = strings.ReplaceAll(menuKeyStr, "Space", " ")
		err = tmpl.Execute(&out, TemplateData{
			Slides:        slides,
			Title:         presentation.Title,
			Menu:          presentation.Menu,
			Port:          port,
			PageNext:      strings.Split(pageNextStr, ","),
			PagePrevious:  strings.Split(pagePreviousStr, ","),
			TerminalCast:  strings.Split(terminalCastStr, ","),
			TerminalClose: strings.Split(terminalCloseStr, ","),
			MenuKey:       strings.Split(menuKeyStr, ","),
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
