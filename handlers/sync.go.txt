package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.com/fer-go/present/data"
)

func sync() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}
		id, chSync := data.Subscribe()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// register and send ID
		se := data.SyncEvent{
			ID:     id,
			Author: 0,
			Slide:  -1000,
		}
		buf, _ := json.Marshal(se)
		fmt.Fprintf(w, "data: %s\n\n", string(buf))
		w.(http.Flusher).Flush()

		var update data.SyncEvent
		for {
			update = <-chSync
			buf, _ = json.Marshal(update)
			fmt.Fprintf(w, "data: %s\n\n", string(buf))
			flusher.Flush()
			if update.Reload {
				break
			}
		}
	}
}

func Sync() http.Handler {
	return AccessControlAllow(sync())
}
