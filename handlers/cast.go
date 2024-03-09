package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/oktalz/present/data"
	"github.com/oktalz/present/data/reader"
	"github.com/oktalz/present/exec"
)

func cast(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	slideStr := r.URL.Query().Get("slide")
	slide, err := strconv.ParseInt(slideStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	presentation := data.Presentation()
	if slide < 0 || slide >= int64(len(presentation)) {
		http.Error(w, "Invalid slide number", http.StatusBadRequest)
	}
	tc := presentation[slide].TerminalCommand
	tcBefore := presentation[slide].TerminalCommandBefore
	tcAfter := presentation[slide].TerminalCommandAfter
	if presentation[slide].CanEdit {
		bodyBytes, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		payload, err := parseJSONData(string(bodyBytes))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(payload.Code)
		for i, code := range payload.Code {
			tc[i].Code.Code = code
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
	cmdResult := exec.CmdStream(tc[0])
	if tcAfter != nil {
		for i := range tcBefore {
			tcAfter[i].Dir = workingDir
			_ = exec.CmdStream(tcAfter[i])
		}
	}
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
		cmdResult[index].Line = strings.ReplaceAll(line.Line, `	`, "    ") // this is not space, but some weird tab
	}
	if maxWidth > 90 {
		maxWidth = 90
	}
	cmd := make([]string, len(tc[0].Cmd))
	copy(cmd, tc[0].Cmd)
	for index, arg := range cmd {
		cmd[index] = strings.ReplaceAll(arg, `"`, `\"`)
	}
	recommendedHeight := maxWidth * 4 / 24
	if lineNum < recommendedHeight {
		lineNum = recommendedHeight
	}
	result := `{"version": 2, "width": ` + strconv.Itoa(maxWidth) + `, "height": ` + strconv.Itoa(lineNum) + `, "timestamp": 17000000000, "env": {"SHELL": "/bin/bash", "TERM": "xterm-256color"}}` + "\n"
	result += `[0.000002, "o", "` + reader.TerminalChar + ` ` + tc[0].App + " " + strings.Join(cmd, " ") + `"]` + "\n"
	result += `[0.010003, "o", "\r\n"]` + "\n"
	for _, line := range cmdResult {
		result += `[` + line.Timestamp + `, "o", "` + line.Line + `\r\n"]` + "\n"
	}
	_, err = w.Write([]byte(result))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Cast() http.Handler {
	return AccessControlAllow(http.HandlerFunc(cast))
}
