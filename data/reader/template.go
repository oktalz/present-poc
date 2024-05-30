package reader

import (
	"bytes"
	"strings"
	"text/template"
)

type TemplateData struct {
	Name string
	Data string
	Vars []string
}

func applyTemplate(fileContent string, templateData TemplateData) string {
	startStr := "." + templateData.Name
	for {
		start := strings.Index(fileContent, startStr)
		if start == -1 {
			break
		}
		start += len(startStr)
		content := fileContent[start:]
		end := strings.Index(content, "\n")
		if end == -1 {
			break
		}
		toReplace := content[:end]
		vars := strings.TrimPrefix(toReplace, " ")
		vars = strings.ReplaceAll(vars, ")", "&#41;")
		vars = strings.ReplaceAll(vars, "(", "&#40;")
		vars = strings.ReplaceAll(vars, "{", "&#123;")
		vars = strings.ReplaceAll(vars, "}", "&#125;")
		vars = strings.ReplaceAll(vars, ".", "&#46;")
		vars = strings.ReplaceAll(vars, "_", "&#95;")
		vars = strings.ReplaceAll(vars, "-", "&#45;")
		vars = strings.ReplaceAll(vars, `"`, "&#34;")
		var data any
		if len(templateData.Vars) == 0 {
			data = vars
		} else {
			templateVars := []string{}
			varsData := strings.Split(vars, " ")
			dataMap := map[string]string{}
			for index := 1; index < len(varsData); index++ {
				if len(templateVars) > index {
					key := strings.TrimPrefix(templateVars[index], ".")
					dataMap[key] = varsData[index]
				}
			}
		}
		tmpl, err := template.New("test").Parse(templateData.Data)
		if err != nil {
			panic(err)
		}

		var tpl bytes.Buffer
		err = tmpl.Execute(&tpl, data)
		if err != nil {
			panic(err)
		}
		result := tpl.String()
		fileContent = strings.Replace(fileContent, startStr+toReplace, result, 1)
	}
	return fileContent
}
