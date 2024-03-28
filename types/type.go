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
	App      string
	Cmd      []string
	Code     Code
	Index    int
	FileName string
	TmpDir   bool
}

type Slide struct {
	Markdown              string            `json:"markdown"`
	Notes                 string            `json:"notes"`
	Terminal              TerminalCommand   `json:"-"`
	Asciinema             *Asciinema        `json:"asciinema,omitempty"`
	Cast                  *Cast             `json:"-"`
	TerminalCommandBefore []TerminalCommand `json:"-"`
	TerminalCommand       []TerminalCommand `json:"-"`
	TerminalCommandAfter  []TerminalCommand `json:"-"`
	UseTmpFolder          bool              `json:"-"`
	CanEdit               bool              `json:"can_edit"`
	HasCast               bool              `json:"cast"`
	HasCastStreamed       bool              `json:"-"`
	HasRun                bool              `json:"run"`
	HasTerminal           bool              `json:"terminal"`
	BackgroundImage       string            `json:"background"`
	BackgroundColor       string            `json:"background_color"`
	PageNumber            int               `json:"page"`
	PageIndex             int               `json:"-"`
	PrintPage             int               `json:"print_page"`
	FontSize              string            `json:"font_size"`
	Html                  string            `json:"html"`
}

type Presentation struct {
	Slides []Slide
	Title  string
}

type TerminalOutputLine struct {
	Timestamp string
	Line      string
}

type ReadOptions struct {
	DevUrl                 string
	DefaultFontSize        string
	DefaultBackgroundColor string
	EveryDashIsACut        bool
}
