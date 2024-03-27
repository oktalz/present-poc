package reader

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gitlab.com/fer-go/present/types"
)

func ReadFiles() []types.Slide {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	ro := types.ReadOptions{
		DevUrl:          `http://localhost:` + port + `/`,
		DefaultFontSize: "3.5vh",
		EveryDashIsACut: false,
	}

	slides, err := listSlideFiles(".")
	if err != nil {
		panic(err)
	}

	var presentationFiles []types.Slide
	lastPageNumber := 0

	var presentationFile []types.Slide
	for _, slide := range slides {
		if len(presentationFiles) > 1 {
			lastPageNumber = presentationFiles[len(presentationFiles)-1].PageNumber
		}
		presentationFile, ro, err = readSlideFile(slide, ro, lastPageNumber)
		if err != nil {
			panic(err)
		}
		presentationFiles = append(presentationFiles, presentationFile...)
	}

	presentations := make([]types.Slide, 0)
	defaultBackend := ""
	var codeBlockShowStart *int
	var codeBlockShowEnd *int
	_ = codeBlockShowStart
	_ = codeBlockShowEnd
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
			for index := 0; index < len(lines); index++ {
				line := lines[index]
				if strings.HasPrefix(line, ".cast.before") {
					newLine := strings.Replace(line, ".cast.before", ".cast.before .", 1)
					lines[index] = newLine
					var tc types.TerminalCommand
					tc, lines = parseCommandBlock(lines, index, nil, nil)
					var c types.Cast
					if slide.Cast != nil {
						c = *slide.Cast
					}
					slide.TerminalCommandBefore = append(slide.TerminalCommandBefore, tc)
					slide.Cast = &c
					slide.HasRun = true
					slide.HasCast = true
					slide.UseTmpFolder = true
					slide.CanEdit = true
					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasCastStream := strings.Contains(slide.Markdown, ".cast.stream")
		if hasCastStream {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".cast.stream") {
					slide.HasCastStreamed = true
					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasCastBlockAfter := strings.Contains(slide.Markdown, ".cast.after")
		if hasCastBlockAfter {
			lines := strings.Split(slide.Markdown, "\n")
			for index := 0; index < len(lines); index++ {
				line := lines[index]
				if strings.HasPrefix(line, ".cast.after") {
					newLine := strings.Replace(line, ".cast.after", ".cast.after .", 1)
					lines[index] = newLine
					var tc types.TerminalCommand
					tc, lines = parseCommandBlock(lines, index, nil, nil)
					var c types.Cast
					if slide.Cast != nil {
						c = *slide.Cast
					}
					slide.TerminalCommandBefore = append(slide.TerminalCommandBefore, tc)
					slide.Cast = &c
					slide.HasRun = true
					slide.HasCast = true
					slide.UseTmpFolder = true
					slide.CanEdit = true
					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasCastBlockShow := strings.Contains(slide.Markdown, ".cast.block.show")
		if hasCastBlockShow {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".cast.block.show") {
					data := strings.Split(line, " ")
					if len(data) < 2 {
						fmt.Println("error parsing: ", line)
					}
					data = strings.Split(data[1], ":")
					if len(data) < 2 {
						fmt.Println("error parsing: ", line)
					}
					start, _ := strconv.Atoi(data[0])
					end, _ := strconv.Atoi(data[1])
					codeBlockShowStart = &start
					codeBlockShowEnd = &end

					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasCastBlockEdit := strings.Contains(slide.Markdown, ".cast.block")
		if hasCastBlockEdit {
			lines := strings.Split(slide.Markdown, "\n")
			fmt.Println("lines", lines)
			//for index, line := range lines {
			for index := 0; index < len(lines); index++ {
				line := lines[index]
				if strings.HasPrefix(line, ".cast.block") {
					var tc types.TerminalCommand
					tc, lines = parseCommandBlock(lines, index, codeBlockShowStart, codeBlockShowEnd)
					var c types.Cast
					if slide.Cast != nil {
						c = *slide.Cast
					}
					slide.TerminalCommand = append(slide.TerminalCommand, tc)
					slide.Cast = &c
					slide.HasRun = true
					slide.HasCast = true
					slide.UseTmpFolder = true
					slide.CanEdit = true
					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					codeBlockShowStart = nil
					codeBlockShowEnd = nil
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
					c := types.Cast{}
					slide.TerminalCommand = append(slide.TerminalCommand, tc)
					slide.Cast = &c
					slide.HasCast = true
					lines = append(lines[:index], lines[index+1:]...)
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
					lines = append(lines[:index], lines[index+1:]...)
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
	shift := 0
	for i := 1; i < len(presentations)-1; i++ {
		if presentations[i].PageNumber < presentations[i-1].PageNumber {
			if presentations[i].PageNumber == 1 {
				shift = presentations[i-1].PageNumber
				_ = shift
			}
		}
	}
	presentations[len(presentations)-1].PrintPage = printPage

	return presentations
}
