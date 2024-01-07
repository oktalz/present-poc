package types

type Asciinema struct {
	Cast     string `json:"cast"`
	URL      string `json:"url"`
	Loop     bool   `json:"loop"`
	AutoPlay bool   `json:"autoplay"`
}

type Cast struct {
	Width           int             `json:"width"`
	Height          int             `json:"height"`
	TerminalCommand TerminalCommand `json:"-"`
}

type TerminalCommand struct {
	Dir        string
	App        string
	Cmd        []string
	CodeHeader string
	Code       string
	CodeFooter string
	FileName   string
	TmpDir     bool
}

type Slide struct {
	Markdown        string          `json:"markdown"`
	Notes           string          `json:"notes"`
	Terminal        TerminalCommand `json:"-"`
	Asciinema       *Asciinema      `json:"asciinema,omitempty"`
	Cast            *Cast           `json:"-"`
	UseTmpFolder    bool            `json:"-"`
	CanEdit         bool            `json:"can_edit"`
	HasCast         bool            `json:"cast"`
	HasRun          bool            `json:"run"`
	HasTerminal     bool            `json:"terminal"`
	BackgroundImage string          `json:"background"`
	PageNumber      int             `json:"page"`
	PrintPage       int             `json:"print_page"`
	FontSize        string          `json:"font_size"`
}

type TerminalOutputLine struct {
	Timestamp string
	Line      string
}

type ReadOptions struct {
	DevUrl          string
	DefaultFontSize string
	EveryDashIsACut bool
}
