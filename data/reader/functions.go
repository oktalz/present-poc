package reader

import (
	"strings"

	"github.com/oktalz/present/types"
)

func parseCommand(command string) types.TerminalCommand {
	parts := strings.Split(command, " ") // TODO handle go run . "some param in quotes" 1 2 ...

	app := parts[2]
	osPath := getOSPath(parts[1])
	tc := types.TerminalCommand{
		Dir: osPath,
		App: app,
		Cmd: parts[3:],
	}
	if osPath == "" {
		tc.FileName = parts[1]
	}
	return tc
}

func parseCommandBlock(lines []string, index int, codeBlockShowStart, codeBlockShowEnd *int) types.TerminalCommand {
	tc := parseCommand(lines[index])
	var codeHeader string
	var code string
	var codeFooter string
	codeStart := index + 2
	codeEnd := len(lines)
	if codeBlockShowStart != nil {
		codeStart += *codeBlockShowStart
	}
	if codeBlockShowEnd != nil {
		codeEnd = index + 2 + *codeBlockShowEnd
	}
	for i := index + 2; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "```") {
			break
		}
		if i < codeStart {
			codeHeader += lines[i] + "\n"
			lines = removeElementFromSlice(lines, i)
			i = i - 1
			codeStart--
			codeEnd--
			continue
		}
		if i > codeEnd {
			codeFooter += lines[i] + "\n"
			lines = removeElementFromSlice(lines, i)
			i = i - 1
			continue
		}
		code += lines[i] + "\n"
	}
	tc.Code = code
	tc.CodeHeader = codeHeader
	tc.CodeFooter = codeFooter
	tc.TmpDir = true
	return tc
}

func replaceWithConditionImage(str, oldStr, newStr, httpStr string) string {
	var sb strings.Builder
	parts := strings.Split(str, oldStr)
	for i, part := range parts {
		if i > 0 && !strings.HasPrefix(part, "http") {
			sb.WriteString(newStr)
		} else if i > 0 {
			sb.WriteString(httpStr)
		}
		sb.WriteString(part)
	}
	return sb.String()
}
