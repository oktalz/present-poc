package reader

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/oktalz/present/types"
)

func listSlideFiles(directory string) ([]string, error) {
	var slideFiles []string

	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".slide") {
			slideFiles = append(slideFiles, filepath.Join(directory, file.Name()))
		}
	}

	return slideFiles, nil
}

func readSlideFile(filename string, ro types.ReadOptions) ([]types.Slide, types.ReadOptions, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, ro, err
	}

	fileContent := string(content)
	fileContent = strings.ReplaceAll(fileContent, ":image(", ":image("+ro.DevUrl)

	lines := strings.Split(fileContent, "\n")
	slides := []types.Slide{}

	var slide strings.Builder
	lastIndex := 1
	templates := map[string]string{}
	shortTemplates := map[string]string{}
	currentFontSize := ro.DefaultFontSize
	defaultEveryDashIsACut := ro.EveryDashIsACut
	_ = defaultEveryDashIsACut
	slideDashCut := ro.EveryDashIsACut
	notes := ""

	for index := 0; index < len(lines); index++ {
		line := lines[index]

		if strings.HasPrefix(line, "...#") {
			// we have a template
			templateLine := strings.TrimPrefix(line, "...#")
			data := strings.SplitN(templateLine, " ", 2)
			shortTemplates[data[0]] = data[1]
			continue
		}
		if strings.HasPrefix(line, ".template") {
			// we have a template
			templateName := strings.TrimPrefix(line, ".template ")
			index++
			var template strings.Builder
			for {
				line = lines[index]
				if strings.HasPrefix(line, ".template.end") {
					break
				}
				template.WriteString(line)
				template.WriteString("\n")
				index++
			}

			templates[templateName] = template.String()
			continue
		}
		if strings.HasPrefix(line, ".notes") {
			// we have notes
			index++
			var notesSB strings.Builder
			for {
				line = lines[index]
				lines[index] = ""
				if strings.HasPrefix(line, ".notes.end") {
					notes = notesSB.String()
					break
				}
				notesSB.WriteString(line)
				notesSB.WriteString("\n")
				index++
			}

			continue
		}
		if strings.HasPrefix(line, ".---") || strings.HasPrefix(line, ".===") {
			// we have reached delimiter, see if we have anything in buffer
			if slide.Len() > 0 {
				slide := slide.String()
				if len(strings.Trim(slide, " \n")) > 0 {
					slides = append(slides, types.Slide{
						Markdown:   slide,
						Notes:      notes,
						PageNumber: lastIndex,
						FontSize:   currentFontSize,
					})
					notes = ""
					lastIndex++
				}
				currentFontSize = ro.DefaultFontSize
			}
			slide.Reset()
			slideDashCut = ro.EveryDashIsACut
			continue
		}
		if strings.HasPrefix(line, ".global.font-size") {
			currentFontSize = strings.TrimPrefix(line, ".global.font-size ")
			ro.DefaultFontSize = currentFontSize
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.dash.is.transition") {
			ro.EveryDashIsACut = true
			slideDashCut = ro.EveryDashIsACut
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.dash.is.transition") {
			slideDashCut = true
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.dash.disable.transition") {
			slideDashCut = false
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.font-size") {
			currentFontSize = strings.TrimPrefix(line, ".slide.font-size ")
			lines[index] = ""
			continue
		}
		isDashCut := slideDashCut && strings.HasPrefix(line, "-")
		if strings.HasPrefix(line, ".cut") || isDashCut {
			// we have reached cut delimiter, see if we have anything in buffer
			converters := strings.Split(line, " ")
			var tmp string
			if slide.Len() > 0 {
				tmp = slide.String()
				slides = append(slides, types.Slide{
					Markdown:   tmp,
					Notes:      notes,
					PageNumber: lastIndex,
					FontSize:   currentFontSize,
				})
				notes = ""
				if !isDashCut {
					for i := 1; i < len(converters); i++ {
						data := strings.SplitN(converters[i], ".", 2)
						// TODO data can be empty, panic if that happens
						tmp = strings.ReplaceAll(tmp, data[0], data[1])
					}
				}
				// lastIndex++ index will remain the same
			}
			slide.Reset()
			slide.WriteString(tmp)
			if isDashCut {
				slide.WriteString(line)
				slide.WriteString("\n")
			}
			continue
		}
		if strings.HasPrefix(line, ".//") {
			// we have reached comment, ignore it
			continue
		}
		isATemplate := false
		if strings.HasPrefix(line, ".") {
			// is it a template?
			for k, v := range templates {
				if strings.HasPrefix(line, "."+k) {
					parts := strings.Split(strings.TrimPrefix(line, "."+k), ".")
					data := map[string]string{}
					for _, p := range parts {
						if p == "" {
							continue
						}
						d := strings.SplitN(p, " ", 2)
						if len(d) != 2 {
							continue
						}
						data[d[0]] = strings.TrimSuffix(d[1], " ")
					}
					tmpl, err := template.New("test").Parse(v)
					if err != nil {
						panic(err)
					}

					var tpl bytes.Buffer
					err = tmpl.Execute(&tpl, data)
					if err != nil {
						panic(err)
					}
					log.Println(tpl.String())
					slide.WriteString(tpl.String())
					slide.WriteString("\n")
					isATemplate = true
					break
				}
			}
		}
		if isATemplate {
			continue
		}
		slide.WriteString(line)
		slide.WriteString("\n")
	}
	if slide.Len() > 0 {
		slides = append(slides, types.Slide{
			Markdown:   slide.String(),
			Notes:      notes,
			PageNumber: lastIndex,
			FontSize:   currentFontSize,
		})
		// notes = ""
		// lastIndex++
	}
	for template, data := range shortTemplates {
		for index := range slides {
			slides[index].Markdown = strings.ReplaceAll(slides[index].Markdown, "...#"+template, data)
		}
	}
	return slides, ro, nil
}

func readFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	return string(content), err
}
