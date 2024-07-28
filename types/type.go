package types

type Code struct {
	Header string
	Code   string
	Footer string
}

type Asciinema struct {
	Cast     string `json:"cast"`
	URL      string `json:"url"`
	Loop     bool   `json:"loop"`
	AutoPlay bool   `json:"autoplay"`
}

type Cast struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type TerminalCommand struct {
	Dir      string
	DirFixed bool
	App      string
	Cmd      []string
	Code     Code
	Index    int
	FileName string
	TmpDir   bool
}

type Slide struct {
	Markdown              string
	Notes                 string
	IsAdmin               bool
	AdminMarkdown         string
	Terminal              TerminalCommand
	Asciinema             *Asciinema
	Cast                  *Cast
	TerminalCommandBefore []TerminalCommand
	TerminalCommand       []TerminalCommand
	TerminalCommandAfter  []TerminalCommand
	UseTmpFolder          bool
	CanEdit               bool
	HasCast               bool
	HasCastStreamed       bool
	HasRun                bool
	HasTerminal           bool
	BackgroundImage       string
	BackgroundColor       string
	PageIndex             int
	PagePrint             int
	FontSize              string
	HTML                  string
	AdminHTML             string
	Link                  string
	LinkNext              string
	LinkPrev              string
	PrintOnly             bool
	PrintDisable          bool
	Title                 string
}

type Menu struct {
	Link      int
	PagePrint int
	PageIndex int
	Title     string
}

type Presentation struct {
	CSS       string
	Slides    []Slide
	Menu      []Menu
	Title     string
	Replacers map[string]string
}

type TerminalOutputLine struct {
	Timestamp string
	Line      string
}

type ReadOptions struct {
	DefaultFontSize        string
	DefaultBackgroundColor string
	EveryDashIsACut        bool
}
