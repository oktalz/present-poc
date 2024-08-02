package parsing

import (
	"strings"
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
