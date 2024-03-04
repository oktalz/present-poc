package reader

import (
	"strings"

	"github.com/oktalz/present/types"
)

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

		hasCastBlockBefore := strings.Contains(slide.Markdown, ".cast.before")
		if hasCastBlockBefore {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".cast.before") {
					newLine := strings.Replace(line, ".cast.before", ".cast.before .", 1)
					lines[index] = newLine
					tc := parseCommandBlock(lines, index)
					var c types.Cast
					if slide.Cast != nil {
						c = *slide.Cast
					}
					c.TerminalCommandBefore = &tc
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

		hasCastBlockAfter := strings.Contains(slide.Markdown, ".cast.after")
		if hasCastBlockAfter {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".cast.after") {
					newLine := strings.Replace(line, ".cast.after", ".cast.after .", 1)
					lines[index] = newLine
					tc := parseCommandBlock(lines, index)
					var c types.Cast
					if slide.Cast != nil {
						c = *slide.Cast
					}
					c.TerminalCommandBefore = &tc
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

		hasCastBlockEdit := strings.Contains(slide.Markdown, ".cast.block.edit")
		if hasCastBlockEdit {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".cast.block.edit") {
					tc := parseCommandBlock(lines, index)
					var c types.Cast
					if slide.Cast != nil {
						c = *slide.Cast
					}
					c.TerminalCommand = tc
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
