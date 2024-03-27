package reader

import (
	"strings"

	"gitlab.com/fer-go/present/types"
)

func parseCommand(command string) types.TerminalCommand {
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
	blockIndex := 0
	_ = blockIndex
	for i := index + 2; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "```") {
			for j := range i {
				if strings.HasPrefix(lines[j], "```") {
					blockIndex++
				}
			}
			tc.Index = blockIndex / 2
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
	tc.Code = types.Code{
		Header: codeHeader,
		Code:   code,
		Footer: codeFooter,
	}
	tc.TmpDir = true
	return tc, lines
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
