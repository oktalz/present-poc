package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gitlab.com/fer-go/present/data"
)

func asciinema(w http.ResponseWriter, r *http.Request) {
	slideStr := r.URL.Query().Get("slide")
	slide, err := strconv.ParseInt(slideStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	presentations := data.Presentation()
	if slide < 0 || slide >= int64(len(presentations)) {
		http.Error(w, "Invalid slide number", http.StatusBadRequest)
	}
	err = json.NewEncoder(w).Encode(presentations[slide].Asciinema)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Asciinema() http.Handler {
	return AccessControlAllow(http.HandlerFunc(asciinema))
}
