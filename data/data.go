package data

import (
	"sync"

	"github.com/oktalz/present/data/reader"
	"github.com/oktalz/present/fsnotify"
	"github.com/oktalz/present/types"
)

var (
	mu           sync.RWMutex
	presentation []types.Slide
	chSyncEvent  map[int]chan SyncEvent
	chUpdate     chan SyncEvent
)

type SyncEvent struct {
	ID     int  `json:"ID"`
	Author int  `json:"Author"`
	Slide  int  `json:"Slide"`
	Reload bool `json:"Reload"`
}

func Presentation() []types.Slide {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]types.Slide, len(presentation))
	copy(result, presentation)
	return result
}

func SetPresentation(p []types.Slide) {
	mu.Lock()
	defer mu.Unlock()
	presentation = p
}

func Subscribe() (int, chan SyncEvent) {
	mu.Lock()
	defer mu.Unlock()
	ch := make(chan SyncEvent)
	id := len(chSyncEvent) + 1
	chSyncEvent[id] = ch
	return id, ch
}

func Init(channelMap map[int]chan SyncEvent, update chan SyncEvent) {
	chSyncEvent = channelMap
	chUpdate = update
	curState := SyncEvent{
		ID:     0,
		Author: 0,
		Slide:  0,
		Reload: true,
	}

	filesModified := fsnotify.FileWatcher()

	// initial read
	go func() {
		filesModified <- struct{}{}
	}()

	go func() {
		for {
			select {
			case <-filesModified:
				mu.Lock()
				presentation = reader.ReadFiles()
				for id, ch := range chSyncEvent {
					ch <- SyncEvent{
						ID:     id,
						Author: 0,
						Reload: true,
					}
				}
				// reset all watchers
				chSyncEvent = make(map[int]chan SyncEvent)
				mu.Unlock()
			case update := <-chUpdate:
				if update.Slide == curState.Slide {
					continue
				}
				curState = update
				mu.RLock()
				for id, ch := range chSyncEvent {
					if id == update.Author {
						continue
					}
					update.ID = id
					ch <- update
				}
				mu.RUnlock()
			}
		}
	}()
}
