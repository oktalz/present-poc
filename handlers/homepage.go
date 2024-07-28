package handlers

import (
	"bytes"
	"cmp"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/oktalz/present-poc/data"
	"github.com/oktalz/present-poc/types"
	"github.com/oktalz/present-poc/ui"
)

type TemplateData struct {
	Slides        []types.Slide
	CSS           string
	Menu          []types.Menu
	Title         string
	Port          int
	PageNext      []string
	PagePrevious  []string
	TerminalCast  []string
	TerminalClose []string
	MenuKey       []string
	Admin         bool
}

func Homepage(port int, userPwd, adminPwd string) http.Handler { //nolint:funlen
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = cookieIDValue(w, r)
		userOK, adminPrivileges := cookieAuth(userPwd, adminPwd, r)
		if userPwd != "" {
			if !(userOK) {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
		}

		presentation := data.Presentation()
		slides := presentation.Slides
		for i := range presentation.Slides {
			slides[i].IsAdmin = adminPrivileges
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
			Admin:         adminPrivileges,
			Slides:        slides,
			CSS:           presentation.CSS,
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
