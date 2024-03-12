package reader

import (
	"bytes"
	"log"
	"strings"
	"text/template"
)

func applyTemplate(lines []string, templateName string, templateVars []string, templateContent string) []string {
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "."+templateName) {
			trim := strings.TrimPrefix(line, "."+templateName)
			trim = strings.TrimPrefix(trim, " ")
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
			log.Println(result)
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

func applyTemplate2(lines []string, templateName string, templateVars []string, templateContent string) []string {
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "."+templateName) {
			parts := strings.Split(strings.TrimPrefix(line, "."+templateName), ".")
			var data any
			dataMap := map[string]string{}
			lastKey := ""
			for _, p := range parts {
				if p == "" {
					continue
				}
				d := strings.SplitN(p, " ", 2)
				if len(d) != 2 {
					continue
				}
				lastKey = d[0]
				dataMap[d[0]] = strings.TrimSuffix(d[1], " ")
			}
			data = dataMap
			if len(dataMap) == 1 && lastKey == "" {
				data = dataMap[lastKey]
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
			log.Println(result)
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
