package reader

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/oklog/ulid/v2"
	"github.com/oktalz/present-poc/markdown"
	"github.com/oktalz/present-poc/types"
)

func ReadFiles() types.Presentation { //nolint:funlen,gocognit,gocyclo,cyclop,maintidx
	ro := types.ReadOptions{
		DefaultFontSize:                "5vh",
		EveryDashIsACut:                false,
		DefaultTerminalFontSize:        "6vh",
		DefaultBackgroundColor:         "black",
		DefaultTerminalFontColor:       "black",
		DefaultTerminalBackgroundColor: "rgb(253, 246, 227)",
	}

	slides, hasHeaderFile, err := listSlideFiles(".")
	if err != nil {
		panic(err)
	}
	headerFile := ""
	if hasHeaderFile {
		headerFileBytes, err := os.ReadFile("_.slide")
		if err != nil {
			panic(err)
		}
		headerFile = string(headerFileBytes)
		headerFile = strings.TrimSpace(headerFile)
		headerFile = strings.ReplaceAll(headerFile, "\r\n", "\n")
	}

	var presentationFiles types.Presentation

	cssBytes, err := os.ReadFile("slide.css")
	if err == nil {
		presentationFiles.CSS = string(cssBytes)
	}

	var presentationFile types.Presentation
	for _, slide := range slides {
		presentationFile, ro, err = readSlideFile(slide, ro, headerFile+"\n")
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

		hasCastBlockBefore := strings.Contains(slide.Markdown, ".cast.parallel")
		if hasCastBlockBefore {
			lines := strings.Split(slide.Markdown, "\n")
			for index := 0; index < len(lines); index++ {
				line := lines[index]
				if strings.HasPrefix(line, ".cast.parallel") {
					// newLine := strings.Replace(line, ".cast.before.dir", ".cast.before.dir .", 1)
					// lines[index] = newLine
					var tc types.TerminalCommand
					tc, lines = parseCommandBlock(lines, index, nil, nil)
					tc.DirFixed = true
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
					index--
				}
			}
		}

		hasCastBlockBefore = strings.Contains(slide.Markdown, ".cast.before")
		if hasCastBlockBefore {
			lines := strings.Split(slide.Markdown, "\n")
			for index := 0; index < len(lines); index++ {
				line := lines[index]
				if strings.HasPrefix(line, ".cast.parallel") {
					// newLine := strings.Replace(line, ".cast.before.dir", ".cast.before.dir .", 1)
					// lines[index] = newLine
					var tc types.TerminalCommand
					tc, lines = parseCommandBlock(lines, index, nil, nil)
					tc.DirFixed = true
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
					index--
				}
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
					index--
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
						continue
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

		hasLink := strings.Contains(slide.Markdown, ".slide.print.disable")
		if hasLink {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.print.disable") {
					slide.PrintDisable = true
					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}
		hasLink = strings.Contains(slide.Markdown, ".slide.print.only")
		if hasLink {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.print.only") {
					slide.PrintOnly = true
					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasLink = strings.Contains(slide.Markdown, ".slide.link.next ")
		if hasLink {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.link.next ") {
					link := strings.TrimPrefix(line, ".slide.link.next ")
					slide.LinkNext = link
					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasLink = strings.Contains(slide.Markdown, ".slide.link.previous ")
		if hasLink {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.link.previous ") {
					link := strings.TrimPrefix(line, ".slide.link.previous ")
					slide.LinkPrev = link
					lines = append(lines[:index], lines[index+1:]...)
					slide.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasLink = strings.Contains(slide.Markdown, ".slide.link ")
		if hasLink {
			lines := strings.Split(slide.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.link ") {
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
					html := `<span onclick="setPageWithUpdate(` + page + `)" style="cursor: pointer;">` + id.String() + `</span>`
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

	// we need to determine what page are for print only
	// or presentation only and align page numbers
	// also if we have print only slides, we need to link slide before and after,
	// the ones that are not print only
	shiftPage := 0
	for index := range len(presentations) {
		presentations[index].PageIndex = index
		if presentations[index].PrintDisable {
			shiftPage++
		}
		presentations[index].PagePrint = index + 1 - shiftPage
		if presentations[index].PrintOnly && index > 0 {
			// find first before, first after and set links (if not set already)
			indexBefore := index - 1
			for indexBefore > 0 {
				if presentations[indexBefore].PrintOnly {
					indexBefore--
				} else {
					break
				}
			}
			indexAfter := index + 1
			for indexAfter < len(presentations)-1 {
				if presentations[indexAfter].PrintOnly {
					indexAfter++
				} else {
					break
				}
			}
			if presentations[indexBefore].Link == "" {
				presentations[indexBefore].Link = ulid.Make().String()
			}
			if presentations[indexAfter].Link == "" {
				presentations[indexAfter].Link = ulid.Make().String()
			}
			presentations[indexBefore].LinkNext = presentations[indexAfter].Link
			presentations[indexAfter].LinkPrev = presentations[indexBefore].Link
		}
	}

	// ok now setup the menu
	menu := make([]types.Menu, 0)
	for i, p := range presentations {
		title := ""
		lines := strings.Split(p.Markdown, "\n")
		for _, line := range lines {
			ldata := line
			ldata = strings.ReplaceAll(ldata, "&#41;", ")")
			ldata = strings.ReplaceAll(ldata, "&#40;", "(")
			ldata = strings.ReplaceAll(ldata, "&#123;", "{")
			ldata = strings.ReplaceAll(ldata, "&#125;", "}")
			ldata = strings.ReplaceAll(ldata, "&#46;", ".")
			ldata = strings.ReplaceAll(ldata, "&#95;", "_")
			ldata = strings.ReplaceAll(ldata, "&#45;", "-")
			ldata = strings.ReplaceAll(ldata, "&#34;", `"`)
			index := strings.LastIndex(ldata, "#")
			if index > -1 {
				title = ldata[index+1:]
				index := strings.LastIndex(title, `"`)
				if index > -1 {
					title = title[index+1:]
				}
				title = strings.Trim(title, ` #*()`)
				break
			}
		}
		if p.Title != "" {
			title = p.Title
		}
		if len(menu) > 0 {
			if menu[len(menu)-1].Title != title {
				menu = append(menu, types.Menu{
					Link:      i,
					PageIndex: p.PageIndex,
					PagePrint: p.PagePrint,
					Title:     title,
				})
			}
		} else {
			menu = append(menu, types.Menu{
				Link:      i,
				PageIndex: p.PageIndex,
				PagePrint: p.PagePrint,
				Title:     title,
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
		CSS:       presentationFiles.CSS,
		Menu:      menu,
		Title:     presentationFiles.Title,
		Replacers: presentationFiles.Replacers,
	}
}
