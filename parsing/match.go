package parsing

import (
	"math"
	"strconv"
	"strings"

	"github.com/oktalz/present-poc/types"
)

func FindData(fileContent, startStr, endStr string) (int, int, string) {
	start := strings.Index(fileContent, startStr)
	if start == -1 {
		return -1, -1, ""
	}
	start += len(startStr)
	content := fileContent[start:]
	index := strings.Index(content, endStr)
	if index == -1 {
		return -1, -1, ""
	}
	result := content[:index]
	// now we need to check if the startStr occures before the endStr
	count := strings.Count(result, startStr)
	if count > 0 {
		// we need to find the last startStr differently, we have nesting
		starts := findAllIndexes(content, startStr)
		ends := findAllIndexes(content, endStr)
		// now I need to find last index, but I need to move as many times as I see startStr within the content
		last := 0
		for i := range starts {
			if starts[i] > ends[last] {
				break
			}
			last++
		}
		result = content[:ends[last]]
	}
	return start, start + len(result), result
}

func FindDataWithAlternative(fileContent, startStr, endStr string, startStr2, endStr2 string) (int, int, string) {
	start := strings.Index(fileContent, startStr)
	if start == -1 {
		start := strings.Index(fileContent, startStr2)
		if start == -1 {
			return -1, -1, ""
		}
		startStr = startStr2
		endStr = endStr2
	}
	start += len(startStr)
	content := fileContent[start:]
	index := strings.Index(content, endStr)
	if index == -1 {
		return -1, -1, ""
	}
	result := content[:index]
	// now we need to check if the startStr occures before the endStr
	count := strings.Count(result, startStr)
	if count > 0 {
		// we need to find the last startStr differently, we have nesting
		starts := findAllIndexes(content, startStr)
		ends := findAllIndexes(content, endStr)
		// now I need to find last index, but I need to move as many times as I see startStr within the content
		last := 0
		for i := range starts {
			if starts[i] > ends[last] {
				break
			}
			last++
		}
		result = content[:ends[last]]
	}
	return start, start + len(result), result
}

func FindDataWithCode(fileContent, startStr, endStr string) (int, int, string, string) {
	start := strings.Index(fileContent, startStr)
	if start == -1 {
		return -1, -1, "", ""
	}
	start += len(startStr)
	content := fileContent[start:]
	index := strings.Index(content, endStr)
	if index == -1 {
		return -1, -1, "", ""
	}
	result := content[:index]
	// now we need to check if the startStr occures before the endStr
	count := strings.Count(result, startStr)
	if count > 0 {
		// we need to find the last startStr differently, we have nesting
		starts := findAllIndexes(content, startStr)
		ends := findAllIndexes(content, endStr)
		// now I need to find last index, but I need to move as many times as I see startStr within the content
		last := 0
		for i := range starts {
			if starts[i] > ends[last] {
				break
			}
			last++
		}
		result = content[:ends[last]]
	}

	// now we need to find the code
	// code is after the result
	codeStart := strings.Index(fileContent[start+len(result)+1:], "```")
	if codeStart == -1 {
		return start, start + len(result), result, ""
	}
	code := fileContent[start+len(result)+1+codeStart+3:]
	codeEnd := strings.Index(code, "```")
	if codeEnd == -1 {
		return -1, -1, "", ""
	}
	codeStart = strings.Index(code, "\n")
	if codeStart == -1 {
		codeStart = 0
	}
	code = code[codeStart+1 : codeEnd]
	return start, start + len(result), result, code
}

func findAllIndexes(text, substring string) []int {
	var indexes []int
	for i := 0; i < len(text); {
		index := strings.Index(text[i:], substring)
		if index == -1 {
			break
		}
		indexes = append(indexes, i+index)
		i += index + len(substring)
	}
	return indexes
}

/*
func ParseCommand(command string) types.TerminalCommand {
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
	if tc.Dir == "" && len(parts) > 1 {
		tc.FileName = parts[1]
	}
	return tc
}

func ParseCommandBlock(lines []string, index int, codeBlockShowStart, codeBlockShowEnd *int) (types.TerminalCommand, []string) {
	tc := ParseCommand(lines[index])
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
*/

type ParseResult struct {
	IsStream           bool
	IsEdit             bool
	Before             []types.TerminalCommand
	Cmd                []types.TerminalCommand
	After              []types.TerminalCommand
	CodeBlockShowStart int
	CodeBlockShowEnd   int
	NewCode            string
}

func ParseCast(cast string, code string) ParseResult {
	//.cast
	// .block
	//+ .stream
	//+ .edit
	//+ .before({folder}go mod init)
	//+ .show(0:8)
	//+ .file(main.go)
	//+ .run(go run .)
	//+ .after({folder}echo "done")
	//+ .parallel()
	//+ .parallel({folder})
	result := ParseResult{
		Before:             []types.TerminalCommand{},
		Cmd:                []types.TerminalCommand{},
		After:              []types.TerminalCommand{},
		CodeBlockShowStart: 0,
		CodeBlockShowEnd:   math.MaxInt,
	}
	result.IsStream = strings.Contains(cast, ".stream")
	result.IsEdit = strings.Contains(cast, ".edit")
	IsBlock := strings.Contains(cast, ".file")
	var start int
	var end int
	var data string
	var content string
	data = cast
	for {
		start, end, content = FindDataWithAlternative(data, ".show(", ")", ".show{", "}")
		if start == -1 {
			break
		}
		parts := strings.Split(content, ":")
		if len(parts) > 1 {
			result.CodeBlockShowStart, _ = strconv.Atoi(parts[0])
			if result.CodeBlockShowStart > 0 {
				result.CodeBlockShowStart-- // to have human readable indexes
			}
			result.CodeBlockShowEnd, _ = strconv.Atoi(parts[1])
		}
		data = data[end+1:]
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data, ".before(", ")", ".before{", "}")
		if start == -1 {
			break
		}
		result.Before = append(result.Before, ParseCommand(content))
		data = data[end+1:]
	}

	hasRun := -1
	data = cast
	for {
		start, _, content = FindDataWithAlternative(data, ".run(", ")", ".run{", "}")
		if start == -1 {
			break
		}
		tc := ParseCommand(content)
		if tc.Dir == "" && !tc.DirFixed {
			tc.TmpDir = true
		}
		if IsBlock {
			splitCode(code, &result, &tc)
		} else {
			tc.Code = types.Code{
				IsEmpty: true,
			}
		}

		result.Cmd = append(result.Cmd, tc)
		hasRun = len(result.Cmd) - 1
		// data = data[end+1:]
		// only one run is allowed per cast
		break //nolint: staticcheck
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data, ".after(", ")", ".after{", "}")
		if start == -1 {
			break
		}
		result.After = append(result.After, ParseCommand(content))
		data = data[end+1:]
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data, ".parallel(", ")", ".parallel{", "}")
		if start == -1 {
			break
		}
		tc := ParseCommand(content)
		result.Before = append(result.Before, tc)
		data = data[end+1:]
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data, ".file(", ")", ".file{", "}")
		if start == -1 {
			break
		}
		if hasRun == -1 {
			tc := types.TerminalCommand{
				FileName: content,
			}
			splitCode(code, &result, &tc)
			result.Cmd = append(result.Cmd, tc)
		} else {
			result.Cmd[hasRun].FileName = content
		}
		// result.FileName = content
		data = data[end+1:]
	}

	return result
}

func ParseCommand(command string) types.TerminalCommand {
	tc := types.TerminalCommand{
		Index: -1,
	}
	start, end, folder := FindData(command, "{", "}")
	if start != -1 {
		tc.Dir = getOSPath(folder)
		tc.DirFixed = true
		command = command[end+1:]
	} else {
		tc.TmpDir = true
	}
	//nolint:godox
	parts := strings.Split(command, " ") // TODO handle go run . "some param in quotes" 1 2 ...
	if len(parts) > 0 {
		tc.App = parts[0]
	}
	if len(parts) > 1 {
		tc.Cmd = parts[1:]
	}
	return tc
}

func splitCode(code string, result *ParseResult, tc *types.TerminalCommand) {
	var Header string
	var Footer string
	codeLines := strings.Split(code, "\n")
	code = ""
	for i := range result.CodeBlockShowStart {
		Header += codeLines[i] + "\n"
	}
	until := result.CodeBlockShowEnd
	if until > len(codeLines) {
		until = len(codeLines)
	}
	for i := result.CodeBlockShowStart; i < until; i++ {
		code += codeLines[i] + "\n"
	}
	for i := result.CodeBlockShowEnd; i < len(codeLines); i++ {
		Footer += codeLines[i] + "\n"
	}
	if code == "\n" {
		code = ""
	}
	if result.CodeBlockShowStart != 0 || result.CodeBlockShowEnd != math.MaxInt {
		if tc.Code.Footer == "" && len(code) > 1 && code[len(code)-2] == '\n' {
			for len(code) > 1 && code[len(code)-2] == '\n' {
				code = code[:len(code)-1]
			}
			result.NewCode = code
		}
	}
	tc.Code = types.Code{
		Header: Header,
		Code:   code,
		Footer: Footer,
	}
	if tc.Code.Header == "" && tc.Code.Code == "" && tc.Code.Footer == "" {
		tc.Code.IsEmpty = true
	}
	if result.CodeBlockShowStart != 0 || result.CodeBlockShowEnd != math.MaxInt {
		result.NewCode = code
	}
}
