package reader

import (
	"strings"

	"github.com/oktalz/present/helper"
	"github.com/oktalz/present/types"
)

func parseCommand(command string) types.TerminalCommand {
	parts := strings.Split(command, " ") // TODO handle go run . "some param in quotes" 1 2 ...

	app := parts[2]
	osPath := getOSPath(parts[1])
	tc := types.TerminalCommand{
		Dir: osPath,
		App: app,
		Cmd: parts[3:],
	}
	if osPath == "" {
		tc.FileName = parts[1]
	}
	return tc
}

func parseCommandBlock(lines []string, index int) types.TerminalCommand {
	tc := parseCommand(lines[index])
	var codeHeader string
	var code string
	var codeFooter string
	for i := index + 2; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "```") {
			break
		}
		if strings.HasPrefix(lines[i], ".HEADER") {
			codeHeader += strings.TrimPrefix(lines[i], ".HEADER") + "\n"
			lines = helper.RemoveElementFromSlice(lines, i)
			i = i - 1
			continue
		}
		if strings.HasPrefix(lines[i], ".FOOTER") {
			codeFooter += strings.TrimPrefix(lines[i], ".FOOTER") + "\n"
			lines = helper.RemoveElementFromSlice(lines, i)
			i = i - 1
			continue
		}
		code += lines[i] + "\n"
	}
	tc.Code = code
	tc.CodeHeader = codeHeader
	tc.CodeFooter = codeFooter
	tc.TmpDir = true
	return tc
}

// filterOmits removes all // OMIT lines from the code
func filterOmits(code string) string {
	// Split the code into lines
	lines := strings.Split(code, "\n")
	result := []string{}

	// Iterate over each line
	for _, line := range lines {
		// Check if the line starts with //OMIT
		if strings.Contains(line, "// OMIT") {
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

func filterOmit(code string, tagList string) string {
	// Split the code into lines
	lines := strings.Split(code, "\n")
	tagLines := map[string][]string{}
	tagsFound := map[string]struct{}{}
	tags := strings.Split(tagList, ",")

	// Iterate over each line
	for i, line := range lines {
		// Check if the line starts with //OMIT
		if strings.Contains(line, "// OMIT") {
			// Get the key after //OMIT
			key := strings.TrimSpace(strings.TrimPrefix(strings.TrimLeft(line, " \t"), "// OMIT"))
			if _, found := tagsFound[key]; found {
				continue
			}
			tagsFound[key] = struct{}{}
			separator := "// OMIT " + key

			// Store the lines into tagLines until the next //OMIT line
			for j := i + 1; j < len(lines); j++ {
				if strings.Contains(lines[j], separator) {
					break
				}
				if strings.Contains(lines[j], "// OMIT") {
					// omit tags are not shown in the presentation
					continue
				}
				tagLines[key] = append(tagLines[key], lines[j])
			}
		}
	}

	result := []string{}
	for _, tag := range tags {
		if len(tagLines[tag]) != 0 {
			result = append(result, strings.Join(tagLines[tag], "\n"))
		}
	}
	if len(result) != 0 {
		return strings.Join(result, "\n")
	}

	if len(tags) == 1 && len(tagLines) != 0 {
		for k := range tagLines {
			return strings.Join(tagLines[k], "\n")
		}
	}
	return code
}

func ReadFiles() []types.Slide {
	ro := types.ReadOptions{
		DevUrl:          `http://localhost:8080/`,
		DefaultFontSize: "3.5vh",
		EveryDashIsACut: false,
	}

	slides, err := listSlideFiles(".")
	if err != nil {
		panic(err)
	}

	var presentationFiles []types.Slide

	var presentationFile []types.Slide
	for _, slide := range slides {
		presentationFile, ro, err = readSlideFile(slide, ro)
		if err != nil {
			panic(err)
		}
		presentationFiles = append(presentationFiles, presentationFile...)
	}

	presentations := make([]types.Slide, 0)
	defaultBackend := ""
	for _, slide := range presentationFiles {
		if defaultBackend != "" {
			slide.BackgroundImage = defaultBackend
		}
		hasDefaultBackground := strings.Contains(slide.Markdown, ".default.background")
		if hasDefaultBackground {
			lines := strings.Split(slide.Markdown, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, ".default.background") {
					slide.Markdown = strings.Replace(slide.Markdown, line, "", 1)
					p := strings.Split(line, " ")
					slide.BackgroundImage = p[1]
					defaultBackend = p[1]
					break
				}
			}
		}
		hasBackground := strings.Contains(slide.Markdown, ".background")
		if hasBackground {
			lines := strings.Split(slide.Markdown, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, ".background") {
					slide.Markdown = strings.Replace(slide.Markdown, line, "", 1)
					p := strings.Split(line, " ")
					slide.BackgroundImage = p[1]
					break
				}
			}
		}

		hasCastBlockEdit := strings.Contains(slide.Markdown, ".cast.block.edit")
		if hasCastBlockEdit {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".cast.block.edit") {
					tc := parseCommandBlock(lines, index)
					c := types.Cast{
						TerminalCommand: tc,
					}
					slide.Cast = &c
					slide.HasRun = true
					slide.HasCast = true
					slide.UseTmpFolder = true
					slide.CanEdit = true
					lines[index] = ""
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasCastBlock := strings.Contains(slide.Markdown, ".cast.block")
		if hasCastBlock {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".cast.block") {
					tc := parseCommandBlock(lines, index)
					c := types.Cast{
						TerminalCommand: tc,
					}
					slide.Cast = &c
					slide.HasRun = true
					slide.HasCast = true
					slide.UseTmpFolder = true
					lines[index] = ""
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasCast := strings.Contains(slide.Markdown, ".cast")
		if hasCast {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".cast") {
					// slide.Markdown = strings.Replace(slide.Markdown, line, "", 1)
					// p := strings.Split(line, " ")
					tc := parseCommand(line)
					c := types.Cast{
						TerminalCommand: tc,
					}
					slide.Cast = &c
					slide.HasCast = true
					lines[index] = ""
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasAsciinema := strings.Contains(slide.Markdown, ".asciinema")
		if hasAsciinema {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".asciinema") {
					// slide.Markdown = strings.Replace(slide.Markdown, line, "", 1)
					p := strings.Split(line, " ")
					ac := types.Asciinema{}
					mov, err := readFile(p[1])
					if err != nil {
						ac.URL = p[1]
					} else {
						ac.Cast = mov
					}
					slide.Asciinema = &ac
					lines[index] = ""
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasSrc := strings.Contains(slide.Markdown, "\n.src")
		if hasSrc {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".src") {
					p := strings.Split(line, " ")
					language := "go"
					codeIndex := 1
					if len(p) > 2 {
						language = p[1]
						codeIndex = 2
					} else {
						// TODO maybe detect language by file extension (map due to .sh -> console)
					}
					code, err := readFile(p[codeIndex])
					if err != nil {
						panic(err)
					}
					code = filterOmits(code)
					code = "``` " + language + "\n" + code + "\n```"
					lines[index] = code
					slide.Markdown = strings.Join(lines, "\n")
				}
			}
		}

		hasTag := strings.Contains(slide.Markdown, "\n.tag")
		if hasTag {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".tag") {
					p := strings.Split(line, " ")
					language := "go"
					// based on extension, detect language
					codeIndex := 1
					//.tag [tag] filepath
					tag := ""
					if len(p) > 2 {
						tag = p[1]
						codeIndex = 2
					}
					code, err := readFile(p[codeIndex])
					if err != nil {
						panic(err)
					}
					code = filterOmit(code, tag)
					code = "``` " + language + "\n" + code + "\n```"
					lines[index] = code
					slide.Markdown = strings.Join(lines, "\n")
				}
			}
		}

		presentations = append(presentations, slide)
	}

	printPage := 1
	for i := 0; i < len(presentations)-1; i++ {
		if presentations[i].PageNumber != presentations[i+1].PageNumber && presentations[i].PrintPage == 0 {
			presentations[i].PrintPage = printPage
			printPage++
		}
	}
	presentations[len(presentations)-1].PrintPage = printPage

	return presentations
}
