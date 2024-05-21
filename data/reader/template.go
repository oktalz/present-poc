package reader

import (
	"bytes"
	"strings"
	"text/template"
)

func applyTemplate(lines []string, templateName string, templateVars []string, templateContent string) []string {
	for i := 0; i < len(lines); i++ { //nolint:varnamelen
		line := lines[i]
		if strings.HasPrefix(line, "."+templateName) { //nolint:nestif
			trim := strings.TrimPrefix(line, "."+templateName)
			trim = strings.TrimPrefix(trim, " ")
			trim = strings.ReplaceAll(trim, ")", "&#41;")
			trim = strings.ReplaceAll(trim, "(", "&#40;")
			trim = strings.ReplaceAll(trim, "{", "&#123;")
			trim = strings.ReplaceAll(trim, "}", "&#125;")
			trim = strings.ReplaceAll(trim, ".", "&#46;")
			trim = strings.ReplaceAll(trim, "_", "&#95;")
			trim = strings.ReplaceAll(trim, "-", "&#45;")
			trim = strings.ReplaceAll(trim, `"`, "&#34;")
			parts := strings.Split(trim, " ")
			var data any
			dataMap := map[string]string{}
			lastKey := ""
			for index, p := range parts {
				if len(templateVars) > index {
					key := strings.TrimPrefix(templateVars[index], ".")
					dataMap[key] = p
					lastKey = key
				}
			}
			data = dataMap
			if len(dataMap) == 1 && lastKey == "" {
				data = dataMap[lastKey]
			}
			if len(dataMap) == 0 && lastKey == "" {
				data = trim
			}
			tmpl, err := template.New("test").Parse(templateContent)
			if err != nil {
				panic(err)
			}

			var tpl bytes.Buffer
			err = tmpl.Execute(&tpl, data)
			if err != nil {
				panic(err)
			}
			result := tpl.String()
			// log.Println(result)
			newLines := strings.Split(result, "\n")
			// Remove the original line and inject newLines at position i
			lines = append(lines[:i], append(newLines, lines[i+1:]...)...)

			// slide.WriteString(tpl.String())
			// slide.WriteString("\n")
			// isATemplate = true
		}
	}
	return lines
}
