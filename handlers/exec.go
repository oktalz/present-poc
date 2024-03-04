package handlers

import (
	"net/http"
	"strconv"

	"github.com/oktalz/present/data"
	"github.com/oktalz/present/exec"
)

func execute(w http.ResponseWriter, r *http.Request) {
	slideStr := r.URL.Query().Get("slide")
	slide, err := strconv.ParseInt(slideStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	presentations := data.Presentation()
	if slide < 0 || slide >= int64(len(presentations)) {
		http.Error(w, "Invalid slide number", http.StatusBadRequest)
	}
	_, err = w.Write(exec.Cmd(presentations[slide].Terminal))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Exec() http.Handler {
	return AccessControlAllow(http.HandlerFunc(execute))
}
