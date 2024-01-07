package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oktalz/present/exec"
	"github.com/oktalz/present/fsnotify"
	"github.com/oktalz/present/reader"
)

type syncEvent struct {
	ID     int  `json:"ID"`
	Author int  `json:"Author"`
	Slide  int  `json:"Slide"`
	Reload bool `json:"Reload"`
}

func main() {
	presentations := reader.ReadFiles()
	mu := sync.RWMutex{}

	filesModified := fsnotify.FileWatcher()
	channels := make(map[int]chan syncEvent)
	chUpdate := make(chan syncEvent)
	curState := syncEvent{
		ID:     0,
		Author: 0,
		Slide:  0,
		Reload: true,
	}
	go func() {
		for {
			select {
			case <-filesModified:
				mu.Lock()
				presentations = reader.ReadFiles()
				for id, ch := range channels {
					ch <- syncEvent{
						ID:     id,
						Author: 0,
						Reload: true,
					}
				}
				// reset all watchers
				channels = make(map[int]chan syncEvent)
				mu.Unlock()
			case update := <-chUpdate:
				if update.Slide == curState.Slide {
					continue
				}
				curState = update
				mu.RLock()
				for id, ch := range channels {
					if id == update.Author {
						continue
					}
					update.ID = id
					ch <- update
					// log.Println("sent update to", id)
				}
				mu.RUnlock()
			}
		}
	}()

	// Handle API endpoint
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		// http: superfluous response.WriteHeader call from main.main.func2 (main.go:185)
		mu.RLock()
		defer mu.RUnlock()
		err := json.NewEncoder(w).Encode(presentations)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Handle API endpoint
	http.HandleFunc("/exec", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		slideStr := r.URL.Query().Get("slide")
		slide, err := strconv.ParseInt(slideStr, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		mu.RLock()
		defer mu.RUnlock()
		if slide < 0 || slide >= int64(len(presentations)) {
			http.Error(w, "Invalid slide number", http.StatusBadRequest)
		}
		_, err = w.Write(exec.Cmd(presentations[slide].Terminal))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Handle cast endpoint
	http.HandleFunc("/cast", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		slideStr := r.URL.Query().Get("slide")
		slide, err := strconv.ParseInt(slideStr, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		mu.RLock()
		defer mu.RUnlock()
		if slide < 0 || slide >= int64(len(presentations)) {
			http.Error(w, "Invalid slide number", http.StatusBadRequest)
		}
		tc := presentations[slide].Cast.TerminalCommand
		if presentations[slide].CanEdit {
			bodyBytes, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tc.Code = string(bodyBytes)
		}
		if tc.TmpDir {
			tmpDir := os.TempDir() + "/present-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			err := os.MkdirAll(tmpDir, 0o755)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			tc.Dir = tmpDir
			defer os.RemoveAll(tmpDir)
			err = os.WriteFile(filepath.Join(tmpDir, tc.FileName), []byte(tc.CodeHeader+tc.Code+tc.CodeFooter), 0o644)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		cmdResult := exec.CmdStream(tc)
		lineNum := len(cmdResult) + 2
		// if lineNum < 12 {
		// 	lineNum = 12
		// }
		// lineNum++
		// lineNum++
		maxWidth := 15
		for index, line := range cmdResult {
			if len(line.Line) > maxWidth {
				maxWidth = len(line.Line)
			}
			cmdResult[index].Line = strings.ReplaceAll(line.Line, `"`, `\"`)
		}
		cmd := make([]string, len(tc.Cmd))
		copy(cmd, tc.Cmd)
		for index, arg := range cmd {
			cmd[index] = strings.ReplaceAll(arg, `"`, `\"`)
		}
		recommendedHeight := maxWidth * 4 / 24
		if lineNum < recommendedHeight {
			lineNum = recommendedHeight
		}
		result := `{"version": 2, "width": ` + strconv.Itoa(maxWidth) + `, "height": ` + strconv.Itoa(lineNum) + `, "timestamp": 17000000000, "env": {"SHELL": "/bin/bash", "TERM": "xterm-256color"}}` + "\n"
		result += `[0.000002, "o", "` + reader.TerminalChar + ` ` + tc.App + " " + strings.Join(cmd, " ") + `"]` + "\n"
		result += `[0.010003, "o", "\r\n"]` + "\n"
		for _, line := range cmdResult {
			result += `[` + line.Timestamp + `, "o", "` + line.Line + `\r\n"]` + "\n"
		}
		_, err = w.Write([]byte(result))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Handle API endpoint
	http.HandleFunc("/asciinema", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		slideStr := r.URL.Query().Get("slide")
		slide, err := strconv.ParseInt(slideStr, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		mu.RLock()
		defer mu.RUnlock()
		if slide < 0 || slide >= int64(len(presentations)) {
			http.Error(w, "Invalid slide number", http.StatusBadRequest)
		}
		err = json.NewEncoder(w).Encode(presentations[slide].Asciinema)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Handle API endpoint
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		body := &syncEvent{}
		if err := json.NewDecoder(r.Body).Decode(body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		body.ID = 0
		body.Reload = false
		chUpdate <- *body
	})

	// Handle API endpoint
	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}
		mu.Lock()
		id := len(channels) + 1
		chSync := make(chan syncEvent)
		channels[id] = chSync
		mu.Unlock()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// register and send ID
		se := syncEvent{
			ID:     id,
			Author: 0,
		}
		buf, _ := json.Marshal(se)
		fmt.Fprintf(w, "data: %s\n\n", string(buf))
		w.(http.Flusher).Flush()

		var update syncEvent
		for {
			update = <-chSync
			buf, _ = json.Marshal(update)
			fmt.Fprintf(w, "data: %s\n\n", string(buf))
			flusher.Flush()
			if update.Reload {
				break
			}
		}
	})

	// Serve static files
	sub, err := fs.Sub(dist, "ui/dist")
	if err != nil {
		panic(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	handler := &fallbackFileServer{
		primary:   http.FileServer(http.FS(sub)),
		secondary: http.FileServer(http.Dir(wd)),
	}
	// http.Handle("/", http.FileServer(http.FS(sub)))
	http.Handle("/", handler)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
