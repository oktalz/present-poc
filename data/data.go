package data

import (
	"log"
	"strings"
	"sync"

	"github.com/oktalz/present-poc/data/reader"
	"github.com/oktalz/present-poc/fsnotify"
	"github.com/oktalz/present-poc/markdown"
	"github.com/oktalz/present-poc/types"
)

var (
	muPresentation sync.RWMutex
	presentation   types.Presentation
)

type Message struct {
	ID     string
	Author string
	Admin  bool
	Msg    []byte
	Slide  int
	Reload bool
	Pool   string
	Value  string
	Data   any
}

func Presentation() types.Presentation {
	muPresentation.RLock()
	defer muPresentation.RUnlock()
	slides := make([]types.Slide, len(presentation.Slides))
	copy(slides, presentation.Slides)
	menu := make([]types.Menu, len(presentation.Menu))
	copy(menu, presentation.Menu)
	result := types.Presentation{
		Slides: slides,
		CSS:    presentation.CSS,
		Menu:   menu,
		Title:  presentation.Title,
	}
	return result
}

func SetPresentation(p types.Presentation) {
	muPresentation.Lock()
	defer muPresentation.Unlock()
	presentation = p
}

func Init(server Server) {
	filesModified := fsnotify.FileWatcher()

	// initial read
	go func() {
		filesModified <- struct{}{}
	}()

	go func() {
		for range filesModified {
			muPresentation.Lock()
			presentation = reader.ReadFiles()
			var err error
			for i := range presentation.Slides {
				var adminHTML string
				if presentation.Slides[i].AdminMarkdown != "" {
					log.Println("AdminMarkdown", presentation.Slides[i].AdminMarkdown)
					adminHTML, err = markdown.Convert(presentation.Slides[i].AdminMarkdown)
					if err != nil {
						log.Println(err)
					}
					log.Println("adminHTML", adminHTML)
					log.Println("AdminMarkdown", presentation.Slides[i].AdminMarkdown)
				}
				res, err := markdown.Convert(presentation.Slides[i].Markdown)
				if err != nil {
					log.Println(err)
				}
				for old, new := range presentation.Replacers {
					res = strings.ReplaceAll(res, old, new)
					if adminHTML != "" {
						adminHTML = strings.ReplaceAll(adminHTML, old, new)
					}
				}
				presentation.Slides[i].HTML = res
				presentation.Slides[i].AdminHTML = adminHTML
			}

			markdown.ResetBlocks()
			server.Broadcast(Message{
				Reload: true,
			})
			muPresentation.Unlock()
		}
	}()
}
