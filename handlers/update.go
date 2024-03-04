package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/oktalz/present/data"
)

func updateHandler(chUpdate chan data.SyncEvent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		body := &data.SyncEvent{}
		if err := json.NewDecoder(r.Body).Decode(body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		body.ID = 0
		body.Reload = false
		chUpdate <- *body
	}
}

func Update(chUpdate chan data.SyncEvent) http.Handler {
	return AccessControlAllow(updateHandler(chUpdate))
}
