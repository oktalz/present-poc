package handlers

import (
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
	slideStr := r.URL.Query().Get("slide")
	slide, err := strconv.ParseInt(slideStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	presentation := data.Presentation()
	if slide < 0 || slide >= int64(len(presentation)) {
		http.Error(w, "Invalid slide number", http.StatusBadRequest)
	}
	tc := presentation[slide].Cast.TerminalCommand
	tcBefore := presentation[slide].Cast.TerminalCommandBefore
	tcAfter := presentation[slide].Cast.TerminalCommandAfter
	if presentation[slide].CanEdit {
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
	if tcBefore != nil {
		tcBefore.Dir = tc.Dir
		_ = exec.CmdStream(*tcBefore)
	}
	cmdResult := exec.CmdStream(tc)
	if tcAfter != nil {
		tcAfter.Dir = tc.Dir
		_ = exec.CmdStream(*tcAfter)
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
}

func Cast() http.Handler {
	return AccessControlAllow(http.HandlerFunc(cast))
}
