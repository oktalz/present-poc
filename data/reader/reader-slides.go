package reader

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/fer-go/present/markdown"
	"gitlab.com/fer-go/present/types"
)

func ReadFiles() types.Presentation { //nolint:funlen,gocognit,gocyclo,cyclop,maintidx
	ro := types.ReadOptions{
		DefaultFontSize: "3.5vh",
		EveryDashIsACut: false,
	}

	slides, err := listSlideFiles(".")
	if err != nil {
		panic(err)
	}

	var presentationFiles types.Presentation
	lastPageNumber := 0

	var presentationFile types.Presentation
	for _, slide := range slides {
		if len(presentationFiles.Slides) > 1 {
			lastPageNumber = presentationFiles.Slides[len(presentationFiles.Slides)-1].PageNumber
		}
		presentationFile, ro, err = readSlideFile(slide, ro, lastPageNumber)
		if err != nil {
			panic(err)
		}
		presentationFiles.Slides = append(presentationFiles.Slides, presentationFile.Slides...)
		if presentationFile.Title != "" {
			presentationFiles.Title = presentationFile.Title
		}
		if presentationFiles.Replacers == nil {
			presentationFiles.Replacers = make(map[string]string)
		}
		for k, v := range presentationFile.Replacers {
			presentationFiles.Replacers[k] = v
		}
	}

	presentations := make([]types.Slide, 0)
	defaultBackend := ""
	var codeBlockShowStart *int
	var codeBlockShowEnd *int
	_ = codeBlockShowStart
	_ = codeBlockShowEnd
	for _, slide := range presentationFiles.Slides {
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
		if hasCastBlockShow { //nolint:nestif
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
			// fmt.Println("lines", lines)
			// for index, line := range lines {
			for index := 0; index < len(lines); index++ {
				line := lines[index]
				if strings.HasPrefix(line, ".cast.block") {
					var terminalCommand types.TerminalCommand
					terminalCommand, lines = parseCommandBlock(lines, index, codeBlockShowStart, codeBlockShowEnd)
					var c types.Cast
					if slide.Cast != nil {
						c = *slide.Cast
					}
					slide.TerminalCommand = append(slide.TerminalCommand, terminalCommand)
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

		hasLink := strings.Contains(slide.Markdown, ".slide.link.next")
		if hasLink {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.link.next") {
					link := strings.TrimPrefix(line, ".slide.link.next ")
					slide.LinkNext = link
					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasLink = strings.Contains(slide.Markdown, ".slide.link.previous")
		if hasLink {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.link.previous") {
					link := strings.TrimPrefix(line, ".slide.link.previous ")
					slide.LinkPrev = link
					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasLink = strings.Contains(slide.Markdown, ".slide.link")
		if hasLink {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.link") {
					link := strings.TrimPrefix(line, ".slide.link ")
					slide.Link = link
					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}
		hasLink = strings.Contains(slide.Markdown, ".link.")
		if hasLink { //nolint:nestif
			lines := strings.Split(slide.Markdown, "\n")
			for index := 0; index < len(lines); index++ {
				line := lines[index]
				if strings.Contains(line, ".link.") {
					strIndex := strings.Index(line, ".link.")
					line := line[strIndex:] //nolint:gocritic
					link := strings.TrimPrefix(line, ".link.")
					strIndex = strings.Index(link, "{")
					if strIndex < 1 {
						continue
					}
					page := link[:strIndex]
					data := link[strIndex+1:]
					strIndex = strings.Index(data, "}")
					if strIndex < 1 {
						continue
					}
					data = data[:strIndex]
					toReplace := `.link.` + page + `{` + data + `}`
					_ = toReplace
					_ = page
					// log.Println(data)

					id := markdown.CreateCleanMD(data)
					html := `<span onclick="setPage(` + page + `)" style="cursor: pointer;">` + id.String() + `</span>`
					lines[index] = strings.Replace(lines[index], toReplace, html, 1)
					_ = link
					_ = index
					// lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					index--
				}
			}
		}

		presentations = append(presentations, slide)
	}

	printPage := 1
	for i := range len(presentations) - 1 {
		if presentations[i].PageNumber != presentations[i+1].PageNumber && presentations[i].PrintPage == 0 {
			presentations[i].PrintPage = printPage
			printPage++
		}
	}
	var shift int
	for i := 1; i < len(presentations)-1; i++ {
		if presentations[i].PageNumber < presentations[i-1].PageNumber {
			if presentations[i].PageNumber == 1 {
				shift = presentations[i-1].PageNumber
				_ = shift
			}
		}
	}
	if len(presentations) > 0 {
		presentations[len(presentations)-1].PrintPage = printPage
	}

	// ok now setup the menu
	menu := make([]types.Menu, 0)
	for i, p := range presentations {
		title := ""
		lines := strings.Split(p.Markdown, "\n")
		for _, line := range lines {
			index := strings.LastIndex(line, "#")
			if index > -1 {
				title = line[index+1:]
				index := strings.LastIndex(title, `"`)
				if index > -1 {
					title = title[index+1:]
				}
				title = strings.Trim(title, ` #*`)
				break
			}
		}
		if len(menu) > 0 {
			if menu[len(menu)-1].Title == title {
				menu[len(menu)-1].Link = i
				menu[len(menu)-1].Page = p.PageNumber
			} else {
				menu = append(menu, types.Menu{
					Link:  i,
					Page:  p.PrintPage,
					Title: title,
				})
			}
		} else {
			menu = append(menu, types.Menu{
				Link:  i,
				Page:  p.PrintPage,
				Title: title,
			})
		}
	}

	links := make(map[string]int, 0)
	for index, p := range presentations {
		if p.Link != "" {
			links[p.Link] = index
		}
	}
	for link, page := range links {
		for index := range len(presentations) {
			p := presentations[index]
			presentations[index].Markdown = strings.ReplaceAll(p.Markdown, link, strconv.Itoa(page))
			if p.LinkNext == link {
				presentations[index].LinkNext = strconv.Itoa(page)
			}
			if p.LinkPrev == link {
				presentations[index].LinkPrev = strconv.Itoa(page)
			}
		}
	}

	return types.Presentation{
		Slides:    presentations,
		Menu:      menu,
		Title:     presentationFiles.Title,
		Replacers: presentationFiles.Replacers,
	}
}
