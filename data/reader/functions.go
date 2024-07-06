package reader

import (
	"strings"

	"github.com/oktalz/present-poc/types"
)

func parseCommand(command string) types.TerminalCommand {
	//nolint:godox
	parts := strings.Split(command, " ") // TODO handle go run . "some param in quotes" 1 2 ...
	tc := types.TerminalCommand{
		Index: -1,
	}
	if len(parts) > 1 {
		tc.Dir = getOSPath(parts[1])
	}
	if len(parts) > 2 {
		tc.App = parts[2]
	}
	if len(parts) > 3 {
		tc.Cmd = parts[3:]
	}
	if tc.Dir == "" {
		tc.FileName = parts[1]
	}
	return tc
}

func parseCommandBlock(lines []string, index int, codeBlockShowStart, codeBlockShowEnd *int) (types.TerminalCommand, []string) {
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
	blockFound := false
	for i := index + 2; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "```") {
			if blockFound {
				break
			}
			blockFound = true
			continue
		}
		if i < codeStart {
			codeHeader += lines[i] + "\n"
			lines = removeElementFromSlice(lines, i)
			i--
			codeStart--
			codeEnd--
			continue
		}
		if i > codeEnd {
			codeFooter += lines[i] + "\n"
			lines = removeElementFromSlice(lines, i)
			i--
			continue
		}
		code += lines[i] + "\n"
	}
	tc.Code = types.Code{
		Header: codeHeader,
		Code:   code,
		Footer: codeFooter,
	}
	tc.TmpDir = true
	return tc, lines
}
