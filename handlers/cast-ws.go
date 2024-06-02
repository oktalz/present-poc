package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gitlab.com/fer-go/present/data"
	"gitlab.com/fer-go/present/exec"
	"nhooyr.io/websocket"
)

func CastWS(server data.Server, adminPwd string) http.Handler { //nolint:funlen,gocognit,revive
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "bye")
		log.Println("connected")

		ch := make(chan string, 10)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		mt, bodyBytes, err := conn.Read(ctx) //nolint:contextcheck // ??? linter is weird
		if err != nil {
			err = conn.Write(context.Background(), mt, []byte(err.Error())) //nolint:contextcheck
			if err != nil {
				log.Println("write:", err)
			}
			return
		}
		defer r.Body.Close()
		payload, err := parseJSONData(string(bodyBytes))
		if err != nil {
			err = conn.Write(context.Background(), mt, []byte(err.Error())) //nolint:contextcheck
			if err != nil {
				log.Println("write:", err)
			}
			return
		}
		if adminPwd != "" && payload.Admin != adminPwd {
			err = conn.Write(context.Background(), mt, []byte("presenter<br>option only<br>🤷 💥 💔<br>")) //nolint:contextcheck
			if err != nil {
				log.Println("write:", err)
			}
			return
		}

		log.Println("recv: " + strconv.Itoa(payload.Slide))
		log.Printf("recv: %s", payload.Code)
		// log.Printf("recv: %s", bodyBytes) - contains pwd
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
		terminalCommand := slides[slide].TerminalCommand
		tcBefore := slides[slide].TerminalCommandBefore
		tcAfter := slides[slide].TerminalCommandAfter
		if slides[slide].CanEdit {
			for i := range terminalCommand {
				if terminalCommand[i].Index == -1 {
					terminalCommand[i].Index = 0
				}
				terminalCommand[i].Code.Code = payload.Code[terminalCommand[i].Index]
			}
		}
		workingDir := os.TempDir() + "/present-" + strconv.FormatInt(time.Now().UnixNano(), 10)
		tmpDirNeeded := false

		for _, cmd := range terminalCommand {
			if cmd.TmpDir { //nolint:nestif
				if !tmpDirNeeded {
					tmpDirNeeded = true
					err = os.MkdirAll(workingDir, 0o755)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
				cmd.Dir = workingDir
				defer os.RemoveAll(workingDir)
				err = os.WriteFile(filepath.Join(workingDir, cmd.FileName), []byte(cmd.Code.Header+cmd.Code.Code+cmd.Code.Footer), 0o600)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			} else {
				workingDir = cmd.Dir
			}
		}

		if len(tcBefore) > 0 {
			for i := range tcBefore {
				if tcBefore[i].DirFixed {
					go exec.CmdStream(tcBefore[i]) //nolint:contextcheck
					time.Sleep(500 * time.Millisecond)
				} else {
					tcBefore[i].Dir = workingDir
					exec.CmdStream(tcBefore[i]) //nolint:contextcheck
				}
			}
		}
		for _, cmd := range terminalCommand {
			cmd.Dir = workingDir
			if cmd.App == "" {
				continue
			}
			go exec.CmdStreamWS(cmd, ch) //nolint:contextcheck
			if slides[slide].HasCastStreamed {
				// this is for streaming
				for line := range ch {
					err = conn.Write(context.Background(), mt, []byte(line)) //nolint:contextcheck
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
				err = conn.Write(context.Background(), mt, []byte(strings.Join(lines, "<br>"))) //nolint:contextcheck
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
			break
		}
		if tcAfter != nil {
			for i := range tcBefore {
				tcAfter[i].Dir = workingDir
				exec.CmdStream(tcAfter[i]) //nolint:contextcheck
			}
		}
	})
}
