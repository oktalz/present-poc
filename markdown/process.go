package markdown

import (
	"strings"

	"github.com/oktalz/present-poc/parsing"
)

func processReplace(fileContent, startStr, endStr string, process func(data string) string) string {
	for {
		start, _, raw := parsing.FindData(fileContent, startStr, endStr)
		if start == -1 {
			return fileContent
		}

		result := process(raw)
		result = CreateCleanRAW(result).String()
		fileContent = strings.ReplaceAll(fileContent, startStr+raw+endStr, result)
	}
}

func processReplaceMiddle(fileContent, startStr, middleStr, endStr string, process func(part1, part2 string) string) string {
	for {
		start, end, part1 := parsing.FindData(fileContent, startStr, middleStr)
		if start == -1 {
			return fileContent
		}
		middle, end, part2 := parsing.FindData(fileContent[end:], middleStr, endStr)
		if middle == -1 {
			return fileContent
		}
		result := process(part1, part2)
		result = CreateCleanRAW(result).String()
		what := startStr + part1 + middleStr + part2 + endStr
		fileContent = strings.ReplaceAll(fileContent, what, result)
	}
}

func processReplaceData(fileContent, str, result string) string {
	start := strings.Index(fileContent, str)
	if start == -1 {
		return fileContent
	}
	result = CreateCleanRAW(result).String()
	fileContent = strings.ReplaceAll(fileContent, str, result)
	return fileContent
}

func processHideData(fileContent, str string) string {
	start := strings.Index(fileContent, str)
	if start == -1 {
		return fileContent
	}
	result := CreateCleanRAW(str).String()
	fileContent = strings.ReplaceAll(fileContent, str, result)
	return fileContent
}
