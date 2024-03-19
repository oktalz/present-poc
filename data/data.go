package data

import (
	"sync"

	"github.com/oklog/ulid/v2"
	"gitlab.com/fer-go/present/data/reader"
	"gitlab.com/fer-go/present/fsnotify"
	"gitlab.com/fer-go/present/types"
)

var (
	muPresentation sync.RWMutex
	presentation   []types.Slide
)

type Message struct {
	ID     ulid.ULID
	Author ulid.ULID
	Msg    []byte
	Slides []types.Slide
	Slide  int
	Reload bool
}

func Presentation() []types.Slide {
	muPresentation.RLock()
	defer muPresentation.RUnlock()
	result := make([]types.Slide, len(presentation))
	copy(result, presentation)
	return result
}

func SetPresentation(p []types.Slide) {
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
		for {
			select {
			case <-filesModified:
				muPresentation.Lock()
				presentation = reader.ReadFiles()
				server.Broadcast(Message{
					Reload: true,
				})
				muPresentation.Unlock()
			}
		}
	}()
}
