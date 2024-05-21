package markdown

import "strings"

func processReplace(fileContent, startStr, endStr string, process func(data string) string) string {
	found := true
	for found {
		found = false
		if strings.Contains(fileContent, startStr) && strings.Contains(fileContent, endStr) {
			start := strings.Index(fileContent, startStr) + len(startStr)
			content := fileContent[start:]
			end := strings.Index(content, endStr)
			raw := content[:end]
			result := process(raw)
			fileContent = strings.ReplaceAll(fileContent, startStr+raw+endStr, result)
			found = true
		}
	}
	return fileContent
}

func processReplaceMiddle(fileContent, startStr, middleStr, endStr string, process func(part1, part2 string) string) string {
	found := true
	for found {
		found = false
		if strings.Contains(fileContent, startStr) && strings.Contains(fileContent, middleStr) && strings.Contains(fileContent, endStr) {
			start := strings.Index(fileContent, startStr) + len(startStr)
			content := fileContent[start:]
			middle := strings.Index(content, middleStr)
			part1 := content[:middle]
			end := strings.Index(content[middle+len(middleStr):], endStr)
			part2 := content[middle+len(middleStr) : end]
			result := process(part1, part2)
			fileContent = strings.ReplaceAll(fileContent, startStr+part1+middleStr+part2+endStr, result)
			found = true
		}
	}
	return fileContent
}

/*
	found := true
	for found {
		found = false
		if strings.Contains(fileContent, ".raw") && strings.Contains(fileContent, ".raw.end") {
			start := strings.Index(fileContent, ".raw") + 4
			end := strings.Index(fileContent, ".raw.end")
			raw := fileContent[start:end]
			id := CreateCleanRAW(raw)
			fileContent = strings.ReplaceAll(fileContent, `.raw`+raw+`.raw.end`, id.String())
			found = true
		}
	}
*/
