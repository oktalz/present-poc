package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gitlab.com/fer-go/present/data"
	"gitlab.com/fer-go/present/exec"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// Allow CORS
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by always returning true
		return true
	},
} // use default options
func castWS(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	log.Println("connected")

	ch := make(chan string, 10)

	mt, bodyBytes, err := c.ReadMessage()
	if err != nil {
		err = c.WriteMessage(mt, []byte(err.Error()))
		if err != nil {
			log.Println("write:", err)
		}
		return
	}
	defer r.Body.Close()
	payload, err := parseJSONData(string(bodyBytes))
	if err != nil {
		err = c.WriteMessage(mt, []byte(err.Error()))
		if err != nil {
			log.Println("write:", err)
		}
		return
	}

	log.Printf("recv: %s", bodyBytes)
	// Simulate sending events (you can replace this with real data)
	// for i := 0; i < 10; i++ {
	// 	fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("Event %d", i))
	// 	time.Sleep(1 * time.Second)
	// 	w.(http.Flusher).Flush()
	// }

	slide := int64(payload.Slide)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	slides := data.Presentation().Slides
	if slide < 0 || slide >= int64(len(slides)) {
		http.Error(w, "Invalid slide number", http.StatusBadRequest)
	}
	tc := slides[slide].TerminalCommand
	tcBefore := slides[slide].TerminalCommandBefore
	tcAfter := slides[slide].TerminalCommandAfter
	if slides[slide].CanEdit {
		for i := range tc {
			tc[i].Code.Code = payload.Code[tc[i].Index]
		}
	}
	workingDir := os.TempDir() + "/present-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	tmpDirNeeded := false

	for _, cmd := range tc {
		if cmd.TmpDir {
			if !tmpDirNeeded {
				tmpDirNeeded = true
				err = os.MkdirAll(workingDir, 0o755)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
			cmd.Dir = workingDir
			defer os.RemoveAll(workingDir)
			err = os.WriteFile(filepath.Join(workingDir, cmd.FileName), []byte(cmd.Code.Header+cmd.Code.Code+cmd.Code.Footer), 0o644)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			workingDir = cmd.Dir
		}
	}

	if len(tcBefore) > 0 {
		for i := range tcBefore {
			tcBefore[i].Dir = workingDir
			_ = exec.CmdStream(tcBefore[i])
		}
	}
	for _, cmd := range tc {
		cmd.Dir = workingDir
		if cmd.App != "" {
			go exec.CmdStreamWS(cmd, ch)
			if slides[slide].HasCastStreamed {
				// this is for streaming
				for line := range ch {
					err = c.WriteMessage(mt, []byte(line))
					if err != nil {
						log.Println("write:", err)
						return
					}
				}
			} else {
				lines := []string{}
				for line := range ch {
					lines = append(lines, line)
				}
				err = c.WriteMessage(mt, []byte(strings.Join(lines, "<br>")))
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
			break
		}

	}
	if tcAfter != nil {
		for i := range tcBefore {
			tcAfter[i].Dir = workingDir
			_ = exec.CmdStream(tcAfter[i])
		}
	}

}

func CastWS() http.Handler {
	return AccessControlAllow(http.HandlerFunc(castWS))
}
