package markdown

import (
	"bytes"
	"log"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
	d2 "github.com/oktalz/goldmark-d2"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
)

type blockData struct {
	ID   ulid.ULID
	Data string
}

var (
	blocks    []blockData
	mdPrivate goldmark.Markdown
	onceMD    sync.Once
)

func GetMD() goldmark.Markdown {
	onceMD.Do(func() {
		mdPrivate = goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithExtensions(&mermaid.Extender{
				NoScript: true,
			}),
			goldmark.WithExtensions(&d2.Extender{
				// Defaults when omitted
				Layout: d2dagrelayout.DefaultLayout,
				// ThemeID: d2themescatalog.,
			}),
			goldmark.WithExtensions(
				emoji.Emoji,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
				// html.WithXHTML(),
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
	res = strings.TrimPrefix(res, "<p>")
	res = strings.TrimSuffix(res, "</p>")

	for index := len(blocks) - 1; index >= 0; index-- {
		res = strings.ReplaceAll(res, blocks[index].ID.String(), blocks[index].Data)
	}
	return res, nil
}

func prepare(md goldmark.Markdown, fileContent string) string { //nolint:funlen,gocognit,maintidx
	fileContent = processReplace(fileContent, ".raw", ".raw.end", func(data string) string {
		return data
	})
	fileContent = processReplace(fileContent, ".raw{", "}", func(data string) string {
		return data
	})
	fileContent = processReplaceMiddle(fileContent, ".api.", "{", "}", func(data, display string) string {
		// .api.pool1{option 2}
		parts := strings.SplitN(data, `.`, 2) //nolint:mnd
		if len(parts) != 2 {
			log.Println("ERROR PARSING", parts)
			return ``
		}
		// log.Println(".api.", parts)
		return `<span onclick="triggerPool('` + parts[0] + `', '` + parts[1] + `')" style="cursor: pointer;">` + CreateCleanMD(prepare(md, display)).String() + `</span>`
	})
	fileContent = processReplace(fileContent, ".center", ".center.end", func(data string) string {
		return `<div style="text-align:center">` + CreateCleanMD(prepare(md, data)).String() + `</div>`
	})
	fileContent = processReplace(fileContent, ".image(", ")", func(data string) string {
		parts := strings.SplitN(data, ` `, 2) //nolint:mnd
		html := `<img src="` + parts[0] + `" `
		width := `auto`
		height := `auto`
		if len(parts) > 1 {
			wh := strings.SplitN(parts[1], `:`, 2)
			if len(wh) == 2 { //nolint:mnd
				if wh[0] != "" {
					width = wh[0]
				}
				if wh[1] != "" {
					height = wh[1]
				}
			}
		}
		html += `style="object-fit: contain; width: ` + width + `; height: ` + height + `;">`
		return html
	})
	fileContent = processReplaceMiddle(fileContent, ".{", "}(", ")", func(style, content string) string {
		id := CreateCleanMD(prepare(md, content))
		html := `<span style="` + style + `">` + id.String() + `</span>`
		return html
	})

	fileContent = processReplace(fileContent, ".bx{", "}", func(data string) string {
		return `<i class='bx ` + data + `'></i>`
	})

	fileContent = processReplace(fileContent, ".table", ".table.end", func(data string) string {
		html := `<table>`
		var currLine int
		trStarted := false
		tdData := ""
		lines := strings.Split(data, "\n")
		i := 0
		for currLine = i + 1; currLine < len(lines); currLine++ {
			if lines[currLine] == ".tr" {
				if trStarted {
					if tdData != "" {
						id := CreateCleanMD(prepare(md, tdData))
						tdData = ""
						html += id.String() + `</td></tr><tr>`
					} else {
						html += `</tr><tr>`
					}
				} else {
					html += `<tr>`
					trStarted = true
				}
				continue
			}
			if strings.HasPrefix(lines[currLine], ".td") {
				line := lines[currLine]
				if tdData != "" {
					id := CreateCleanMD(prepare(md, tdData))
					// solution := prepare(md, tdData)
					html += id.String() + `</td><td>`
				} else {
					html += `<td>`
				}
				parts := strings.Split(line, " ")
				if len(parts) > 1 && strings.Join(parts[1:], " ") != "" {
					id := CreateCleanMD(strings.Join(parts[1:], " "))
					// solution := prepare(md, strings.Join(parts[1:], " "))
					html += id.String() + `</td>`
				}
				tdData = ""
				continue
			}
			tdData += lines[currLine] + "\n"
		}
		id := CreateCleanMD(prepare(md, tdData))
		// solution := prepare(md, tdData)
		if trStarted {
			html += id.String() + `</td></tr></table>`
		} else {
			html += `</table>`
		}

		return html
	})

	lines := strings.Split(fileContent, "\n")
	for i := 0; i < len(lines); i++ {
		if i >= len(lines) {
			break
		}
		index := strings.Index(lines[i], ".style ")
		isStyleBlock := false
		for index != -1 {
			lines[i], isStyleBlock = convertStyle(lines[i])
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
		if strings.HasPrefix(lines[i], ".slide.enable.overflow") {
			centerLines := lines[i+1:]
			var buf bytes.Buffer
			for index, line := range centerLines {
				if index > 0 {
					buf.WriteString("\n")
				}
				buf.WriteString(line)
			}
			id := CreateCleanMD(prepare(md, buf.String()))
			// solution := prepare(md, buf.String())
			lines[i] = `<div class="box-overflow">` + id.String() + `</div>`
			lines = lines[:i+1]
		}
		if strings.HasPrefix(lines[i], ".graph.") {
			id := ulid.Make().String()
			content := strings.TrimPrefix(lines[i], ".graph.")
			data := strings.Split(content, ".")
			graphType := " graph-pie"
			if len(data) > 1 && data[1] == "bar" {
				graphType = " graph-bar"
			}
			if len(data) > 0 {
				lines[i] = `<pre class="mermaid graph-` + data[0] + graphType + `" id="dynamicGraph-` + id + `">flowchart LR;    &nbsp;</pre>`
			} else {
				// log error ?
			}
		}
		if strings.HasPrefix(lines[i], ".tab") { //nolint:nestif
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
						tabContent = "" //nolint:wastedassign
					}
					lines[currLine] = `</div><div class="tabs">`
					// time to create footers
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
				currLine--
			}
			_ = tabs
		}
	}
	fileContent = strings.Join(lines, "\n")

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

func CreateCleanRAW(data string) ulid.ULID {
	id := ulid.Make()
	blocks = append(blocks, blockData{ID: id, Data: data})
	return id
}

func convertStyle(line string) (result string, isBlock bool) { //nolint:nonamedreturns
	index := strings.Index(line, ".style ")
	partBefore := line[:index] //nolint:gocritic
	partStyle := line[index:]
	partAfter := ""
	index = strings.Index(partStyle[1:], ".style ")
	if index != -1 {
		partAfter = partStyle[:index+1]
		partStyle = partStyle[index+1:]
	}
	parts := strings.SplitN(partStyle, `"`, 3)
	// log.Println(parts)
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
