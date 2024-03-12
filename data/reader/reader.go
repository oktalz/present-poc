package reader

import (
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/fer-go/present/types"
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
	fileContent = strings.ReplaceAll(fileContent, ".center.end", ":::")
	fileContent = strings.ReplaceAll(fileContent, ".center", "::: center")
	fileContent = replaceWithConditionImage(fileContent, ".image(", ":image("+ro.DevUrl, ":image(")

	lines := strings.Split(fileContent, "\n")
	slides := []types.Slide{}

	var slide strings.Builder
	lastIndex := 1
	replacers := map[string]string{}
	currentFontSize := ro.DefaultFontSize
	currentBackgroundColor := ro.DefaultBackgroundColor
	defaultEveryDashIsACut := ro.EveryDashIsACut
	_ = defaultEveryDashIsACut
	slideDashCut := ro.EveryDashIsACut
	notes := ""

	for index := 0; index < len(lines); index++ {
		line := lines[index]
		if strings.HasPrefix(line, ".template") {
			// we have a template
			templateName := strings.TrimPrefix(line, ".template ")
			lines[index] = ""
			index++
			var template strings.Builder
			for {
				line = lines[index]
				lines[index] = ""
				if strings.HasPrefix(line, ".template.end") {
					break
				}
				template.WriteString(line)
				template.WriteString("\n")
				index++
			}
			lines = applyTemplate(lines, templateName, strings.TrimSuffix(template.String(), "\n"))
			continue
		}
	}

	for index := 0; index < len(lines); index++ {
		line := lines[index]

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
						Markdown:        slide,
						Notes:           notes,
						PageNumber:      lastIndex,
						FontSize:        currentFontSize,
						BackgroundColor: currentBackgroundColor,
					})
					notes = ""
					lastIndex++
				}
				currentFontSize = ro.DefaultFontSize
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
		})
		// notes = ""
		// lastIndex++
	}
	for pattern, data := range replacers {
		for index := range slides {
			slides[index].Markdown = strings.ReplaceAll(slides[index].Markdown, pattern, data)
		}
	}
	return slides, ro, nil
}

func readFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	return string(content), err
}
