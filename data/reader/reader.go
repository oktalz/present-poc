package reader

import (
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/fer-go/present/types"
)

func listSlideFiles(directory string) ([]string, bool, error) {
	var slideFiles []string
	hasHeaderFile := false

	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, hasHeaderFile, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".slide") {
			if file.Name() == "_.slide" {
				hasHeaderFile = true
				continue
			}
			slideFiles = append(slideFiles, filepath.Join(directory, file.Name()))
		}
	}

	return slideFiles, hasHeaderFile, nil
}

func readSlideFile(filename string, ro types.ReadOptions, lastPageNumber int, headerFile string) (types.Presentation, types.ReadOptions, error) { //nolint:funlen,gocognit,maintidx
	title := ""
	content, err := os.ReadFile(filename)
	if err != nil {
		return types.Presentation{}, ro, err
	}

	fileContent := headerFile + string(content)
	slides := []types.Slide{}

	var slide strings.Builder
	lastIndex := 1 + lastPageNumber
	replacers := map[string]string{}
	replacersAfter := map[string]string{}
	currentSlideTitle := ""
	currentFontSize := ro.DefaultFontSize
	currentBackgroundColor := ro.DefaultBackgroundColor
	defaultEveryDashIsACut := ro.EveryDashIsACut
	_ = defaultEveryDashIsACut
	slideDashCut := ro.EveryDashIsACut
	notes := ""

	startStr := ".template "
	endStr := ".template.end"
	templates := []TemplateData{}
	for {
		start := strings.Index(fileContent, startStr)
		if start == -1 {
			break
		}
		start += len(startStr)
		content := fileContent[start:]
		end := strings.Index(content, endStr)
		if end == -1 {
			break
		}
		raw := content[:end]
		data := strings.Split(raw, "\n")
		templateVars := []string{}
		dataVars := strings.Split(data[0], " ")
		for i := 1; i < len(dataVars); i++ {
			templateVars = append(templateVars, dataVars[i])
		}
		templates = append(templates, TemplateData{
			Name: strings.Split(data[0], " ")[0],
			Data: strings.Join(data[1:], "\n"),
			Vars: templateVars,
		})
		fileContent = strings.ReplaceAll(fileContent, startStr+raw+endStr, "")
	}
	for i := len(templates) - 1; i >= 0; i-- {
		fileContent = applyTemplate(fileContent, templates[i])
	}

	lines := strings.Split(fileContent, "\n")

	for index := 0; index < len(lines); index++ {
		line := lines[index]

		if strings.HasPrefix(line, ".replace.after") {
			// we have a .replace.after line
			templateLine := strings.TrimPrefix(line, ".replace.after ")
			data := strings.SplitN(templateLine, " ", 2)
			// if replacers[data[0]] == "" { // replacing original is allowed
			if len(data) > 1 {
				replacersAfter[data[0]] = data[1]
			} else {
				replacersAfter[data[0]] = ""
			}
			// }
			continue
		}
		if strings.HasPrefix(line, ".replace") {
			// we have a .replace line
			templateLine := strings.TrimPrefix(line, ".replace ")
			data := strings.SplitN(templateLine, " ", 2)
			// if replacers[data[0]] == "" { // replacing original is allowed
			if len(data) > 1 {
				replacers[data[0]] = data[1]
			} else {
				replacers[data[0]] = ""
			}
			// }
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
				notesSB.WriteString("<br>")
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
						Markdown:        slide,
						Notes:           notes,
						PageNumber:      lastIndex,
						FontSize:        currentFontSize,
						BackgroundColor: currentBackgroundColor,
						Title:           currentSlideTitle,
					})
					notes = ""
					lastIndex++
				}
				currentFontSize = ro.DefaultFontSize
				currentSlideTitle = ""
				currentBackgroundColor = ro.DefaultBackgroundColor
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
		if strings.HasPrefix(line, ".global.background-color") {
			currentBackgroundColor = strings.TrimPrefix(line, ".global.background-color ")
			ro.DefaultBackgroundColor = currentBackgroundColor
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
		if strings.HasPrefix(line, ".slide.title") {
			currentSlideTitle = strings.TrimPrefix(line, ".slide.title ")
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".title") {
			title = strings.TrimPrefix(line, ".title ")
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.background-color") {
			currentBackgroundColor = strings.TrimPrefix(line, ".slide.background-color ")
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
					Markdown:        tmp,
					Notes:           notes,
					PageNumber:      lastIndex,
					FontSize:        currentFontSize,
					BackgroundColor: currentBackgroundColor,
					Title:           currentSlideTitle,
				})
				notes = ""
				if !isDashCut {
					for i := 1; i < len(converters); i++ {
						data := strings.SplitN(converters[i], ".", 2)
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
		slide.WriteString(line)
		slide.WriteString("\n")
	}
	if slide.Len() > 0 {
		slides = append(slides, types.Slide{
			Markdown:        slide.String(),
			Notes:           notes,
			PageNumber:      lastIndex,
			FontSize:        currentFontSize,
			BackgroundColor: currentBackgroundColor,
			Title:           currentSlideTitle,
		})
		// notes = ""
		// lastIndex++
	}
	for pattern, data := range replacers {
		for index := range slides {
			slides[index].Markdown = strings.ReplaceAll(slides[index].Markdown, pattern, data)
		}
	}
	return types.Presentation{
		Slides:    slides,
		Title:     title,
		Replacers: replacersAfter,
	}, ro, nil
}

func readFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	return string(content), err
}
