package markdown

import (
	"bytes"
	"log"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
)

type blockData struct {
	ID   ulid.ULID
	Data string
}

var blocks []blockData
var mdPrivate goldmark.Markdown
var onceMD sync.Once

func GetMD() goldmark.Markdown {
	onceMD.Do(func() {
		mdPrivate = goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithExtensions(&mermaid.Extender{
				NoScript: true,
			}),
			goldmark.WithExtensions(
				emoji.Emoji,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
				//html.WithXHTML(),
				html.WithUnsafe(),
			),
		)
	})
	return mdPrivate
}

func ResetBlocks() {
	blocks = nil
}

func Convert(source string) (string, error) {
	md := GetMD()
	var buf bytes.Buffer
	if err := md.Convert([]byte(prepare(md, source)), &buf); err != nil {
		return "", err
	}
	res := buf.String()
	for index := len(blocks) - 1; index >= 0; index-- {
		res = strings.ReplaceAll(res, blocks[index].ID.String(), blocks[index].Data)
	}
	return res, nil
}

func prepare(md goldmark.Markdown, fileContent string) string {
	lines := strings.Split(fileContent, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], ".image(") || strings.Contains(lines[i], ":image(") {
			//:image(http://localhost:8080/assets/images/3.png :50vh)
			//<div style="position: absolute; top: 35vh; left: 15vh; transform: rotate(-15deg);"><p><img src="http://localhost:8080/assets/images/3.png" style=" object-fit: contain; width: auto; height: 50vh;" "=""></p></div>
			imageType := `.image(`
			dotImageIndex := strings.Index(lines[i], ".image(")
			if dotImageIndex == -1 {
				dotImageIndex = strings.Index(lines[i], `:image(`)
				imageType = `:image(`
			}
			if dotImageIndex == -1 {
				continue
			}
			data := lines[i][dotImageIndex:]

			dotImageIndex = strings.Index(data, ")")
			if dotImageIndex == -1 {
				continue
			}
			data = data[7:dotImageIndex]

			parts := strings.SplitN(data, ` `, 2)
			//log.Println(parts)
			_ = parts
			html := `<img src="` + parts[0] + `" `
			width := `auto`
			height := `auto`

			if len(parts) > 1 {
				wh := strings.SplitN(parts[1], `:`, 2)
				if len(wh) == 2 {
					if wh[0] != "" {
						width = wh[0]
					}
					if wh[1] != "" {
						height = wh[1]
					}
				}
			}
			html += `style="object-fit: contain; width: ` + width + `; height: ` + height + `;">`
			data = imageType + data + ")"
			data = strings.ReplaceAll(lines[i], data, html)
			lines[i] = data
		}
		if strings.Contains(lines[i], "{") && strings.Contains(lines[i], "}(") {
			//{red}(to text)
			//<span id="md-convert" style="color: red;">to text</span>

			start := strings.Index(lines[i], "{") + 1
			end := strings.Index(lines[i], "}(")
			color := lines[i][start:end]
			data := lines[i][end+2:]
			end2 := strings.Index(data, ")")
			data = data[:end2]
			id := CreateCleanMD(data)
			html := `<span style="color: ` + color + `;">` + id.String() + `</span>`
			data = strings.ReplaceAll(lines[i], `{`+color+`}(`+data+`)`, html)
			lines[i] = data
			i--
		}
	}
	for i := 0; i < len(lines); i++ {
		if i >= len(lines) {
			break
		}
		index := strings.Index(lines[i], ".style ")
		isStyleBlock := false
		for index != -1 {
			lines[i], isStyleBlock = convertStyle(md, lines[i])
			index = strings.Index(lines[i], ".style ")
			if index != -1 {
				log.Println(lines[i])
			}
		}
		if isStyleBlock {
			var endLine int
			for endLine = i + 1; endLine < len(lines); endLine++ {
				if lines[endLine] == ".style.end" {
					break
				}
			}
			if endLine > len(lines) {
				endLine = len(lines) - 1
			}
			var centerLines []string
			for j := i + 1; j < endLine; j++ {
				centerLines = append(centerLines, lines[j])
			}
			var buf bytes.Buffer
			for index, line := range centerLines {
				if index > 0 {
					buf.WriteString("\n")
				}
				buf.WriteString(line)
			}
			solution := prepare(md, buf.String())
			id := CreateCleanMD(solution)
			lines[i] = lines[i] + "\n" + id.String() + `</div>`
			lines = append(lines[:i+1], lines[endLine+1:]...)
		}
		//if strings.HasPrefix(lines[i], ".style ") {
		// if strings.Contains(lines[i], ".style ") {
		// 	//.style "font-size: 3.8em;text-shadow: 0 0 3px #FFFFFF, 0 0 15px #000000;" Present tool: Go + VUE
		// 	// <div style="font-size: 3.8em;text-shadow: 0 0 3px #FFFFFF, 0 0 15px #000000;"><p><strong>Present tool: Go + VUE</strong></p>

		// 	parts := strings.SplitN(lines[i], `"`, 3)
		// 	log.Println(parts)
		// 	_ = parts
		// 	if len(parts) == 1 {
		// 		continue
		// 	}
		// 	if len(parts) == 3 {
		// 		solution := createCleanMD(md, parts[2])
		// 		lines[i] = `<div style="` + parts[1] + `">` + solution + `</div>`
		// 	}
		// 	if len(parts) == 2 {
		// 		lines[i] = `<div style="` + parts[1] + `">`
		// 	}
		// }
		if strings.HasPrefix(lines[i], ".table") {
			var currLine int
			lines[i] = `<table>`
			trStarted := false
			tdStarted := false
			tdData := ""
			for currLine = i + 1; currLine < len(lines); currLine++ {
				if lines[currLine] == ".table.end" {
					id := CreateCleanMD(tdData)
					//solution := prepare(md, tdData)
					if trStarted {
						lines[currLine] = id.String() + `</td></tr></table>`
					} else {
						lines[currLine] = `</table>`
					}
					trStarted = false
					break
				}
				if lines[currLine] == ".tr" {
					if trStarted {
						if tdData != "" {
							id := CreateCleanMD(tdData)
							//solution := prepare(md, tdData)
							tdData = ""
							tdStarted = false
							lines[currLine] = id.String() + `</td></tr><tr>`
						} else {
							lines[currLine] = `</tr><tr>`
						}
					} else {
						lines[currLine] = `<tr>`
						trStarted = true
					}
				}
				if strings.HasPrefix(lines[currLine], ".td") {
					tdStarted = false
					line := lines[currLine]
					if tdData != "" {
						id := CreateCleanMD(tdData)
						//solution := prepare(md, tdData)
						lines[currLine] = id.String() + `</td><td>`
					} else {
						lines[currLine] = `<td>`
					}
					parts := strings.Split(line, " ")
					if len(parts) > 1 && strings.Join(parts[1:], " ") != "" {
						id := CreateCleanMD(strings.Join(parts[1:], " "))
						//solution := prepare(md, strings.Join(parts[1:], " "))
						lines[currLine] += id.String() + `</td>`
					} else {
						tdStarted = true
					}
					tdData = ""
					continue
				}
				if tdStarted {
					tdData += "\n" + lines[currLine]
					lines = append(lines[:currLine], lines[currLine+1:]...)
					currLine -= 1
				}
			}
		}
		if strings.HasPrefix(lines[i], ".center") {
			// first find the .center.end line
			var endLine int
			for endLine = i + 1; endLine < len(lines); endLine++ {
				if lines[endLine] == ".center.end" {
					break
				}
			}
			if endLine > len(lines) {
				endLine = len(lines) - 1
			}

			var centerLines []string
			for j := i + 1; j < endLine; j++ {
				centerLines = append(centerLines, lines[j])
			}
			var buf bytes.Buffer
			for index, line := range centerLines {
				if index > 0 {
					buf.WriteString("\n")
				}
				buf.WriteString(line)
			}
			id := CreateCleanMD(prepare(md, buf.String()))
			//solution := prepare(md, buf.String())
			lines[i] = `<div style="text-align:center">` + id.String() + `</div>`
			lines = append(lines[:i+1], lines[endLine+1:]...)
		}
		if strings.HasPrefix(lines[i], ".tab") {
			var currLine int
			lines[i] = `<div class="tab">`
			tabContent := ""
			activeID := ""
			currentTabID := ""
			tabs := []string{}
			tabsID := []string{}
			for currLine = i + 1; currLine < len(lines); currLine++ {
				if lines[currLine] == ".tabs.end" {
					if tabContent != "" {
						tabs = append(tabs, tabContent)
						tabsID = append(tabsID, currentTabID)
						tabContent = ""
					}
					lines[currLine] = `</div><div class="tabs">`
					//time to create footers
					for index, data := range tabs {
						contentID := CreateCleanMD(prepare(md, data))
						class := " hidden-tab"
						id := tabsID[index]
						if id == activeID {
							class = ""
						}
						lines[currLine] += `<div class="tabcontent` + class + `" id="` + id + `">` + contentID.String() + `</div>`
					}
					lines[currLine] += `</div>`
					break
				}
				if strings.HasPrefix(lines[currLine], ".tab") {
					if tabContent != "" {
						tabs = append(tabs, tabContent)
						tabsID = append(tabsID, currentTabID)
						tabContent = ""
					}
					parts := strings.Split(lines[currLine], " ")
					tabName := parts[1]
					tabActive := ""
					currentTabID = ulid.Make().String()
					if strings.HasSuffix(parts[0], ".active") {
						tabActive = " active"
						activeID = currentTabID
					}
					lines[currLine] = `<button class="tablinks` + tabActive + `" onclick="tabChangeGlobal('` + currentTabID + `')" id='tab-` + currentTabID + `'>` + tabName + `</button>`
					continue
				}
				tabContent += "\n" + lines[currLine]
				lines = append(lines[:currLine], lines[currLine+1:]...)
				currLine -= 1
			}
			_ = tabs
		}

		// 	}
		// 	//id := createCleanMD(md, prepare(md, buf.String()))
		// 	//solution := prepare(md, buf.String())
		// 	//lines[i] = `<div style="text-align:center">` + id.String() + `</div>`
		// 	//lines = append(lines[:i+1], lines[endLine+1:]...)
		// }
	}
	fileContent = strings.Join(lines, "\n")

	// TODO: remove this
	fileContent = strings.ReplaceAll(fileContent, ".style.end", `</div>`)
	fileContent = strings.ReplaceAll(fileContent, "____________", `<hr>`)
	return fileContent
}

func CreateCleanMD(data string) ulid.ULID {
	md := GetMD()
	var buf bytes.Buffer
	id := ulid.Make()
	if err := md.Convert([]byte(data), &buf); err != nil {
		blocks = append(blocks, blockData{ID: id, Data: ""})
		return id
	}
	solution := strings.TrimPrefix(buf.String(), "<p>")
	solution = strings.TrimSuffix(solution, "\n")
	solution = strings.TrimSuffix(solution, "</p>")

	blocks = append(blocks, blockData{ID: id, Data: solution})
	return id
}

func convertStyle(md goldmark.Markdown, line string) (result string, isBlock bool) {
	index := strings.Index(line, ".style ")
	partBefore := line[:index]
	partStyle := line[index:]
	partAfter := ""
	index = strings.Index(partStyle[1:], ".style ")
	if index != -1 {
		partAfter = partStyle[:index+1]
		partStyle = partStyle[index+1:]
	}
	parts := strings.SplitN(partStyle, `"`, 3)
	//log.Println(parts)
	_ = parts
	if len(parts) == 1 || len(parts) == 3 && parts[2] == "" {
		parts := strings.SplitN(partStyle, ` `, 2)
		parts[1] = strings.TrimPrefix(parts[1], `"`)
		parts[1] = strings.TrimSuffix(parts[1], `"`)
		line = partBefore + `<div style='` + parts[1] + `'>` + partAfter
		return line, true
	}
	if len(parts) == 3 {
		id := CreateCleanMD(parts[2])
		line = partBefore + `<div style='` + parts[1] + `'>` + id.String() + `</div>` + partAfter
	}
	if len(parts) == 2 {
		line = partBefore + `<div style='` + parts[1] + `'>` + partAfter
	}
	return line, false
}
