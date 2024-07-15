package markdown

import "strings"

func processReplace(fileContent, startStr, endStr string, process func(data string) string) string {
	for {
		start := strings.Index(fileContent, startStr)
		if start == -1 {
			return fileContent
		}
		start += len(startStr)
		content := fileContent[start:]
		end := strings.Index(content, endStr)
		if end == -1 {
			return fileContent
		}
		raw := content[:end]
		result := process(raw)
		result = CreateCleanRAW(result).String()
		fileContent = strings.ReplaceAll(fileContent, startStr+raw+endStr, result)
	}
}

func processReplaceMiddle(fileContent, startStr, middleStr, endStr string, process func(part1, part2 string) string) string {
	for {
		start := strings.Index(fileContent, startStr)
		if start == -1 {
			return fileContent
		}
		start += len(startStr)
		content := fileContent[start:]
		middle := strings.Index(content, middleStr)
		if middle == -1 {
			return fileContent
		}
		part1 := content[:middle]
		content = content[middle+len(middleStr):]
		end := strings.Index(content, endStr)
		if end == -1 {
			return fileContent
		}
		part2 := content[:end]
		result := process(part1, part2)
		result = CreateCleanRAW(result).String()
		what := startStr + part1 + middleStr + part2 + endStr
		fileContent = strings.ReplaceAll(fileContent, what, result)
	}
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
