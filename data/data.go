package data

import (
	"log"
	"sync"

	"github.com/oklog/ulid/v2"
	"gitlab.com/fer-go/present/data/reader"
	"gitlab.com/fer-go/present/fsnotify"
	"gitlab.com/fer-go/present/markdown"
	"gitlab.com/fer-go/present/types"
)

var (
	muPresentation sync.RWMutex
	presentation   types.Presentation
)

type Message struct {
	ID     ulid.ULID
	Author ulid.ULID
	Msg    []byte
	Slide  int
	Reload bool
}

func Presentation() types.Presentation {
	muPresentation.RLock()
	defer muPresentation.RUnlock()
	slides := make([]types.Slide, len(presentation.Slides))
	copy(slides, presentation.Slides)
	result := types.Presentation{
		Slides: slides,
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
			for i := range presentation.Slides {
				res, err := markdown.Convert(presentation.Slides[i].Markdown)
				if err != nil {
					log.Println(err)
				}
				presentation.Slides[i].Html = res
			}
			server.Broadcast(Message{
				Reload: true,
			})
			muPresentation.Unlock()
		}
	}()
}
