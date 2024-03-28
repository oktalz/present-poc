package ui

import _ "embed"

//go:embed web.tmpl
var WebTemplate []byte

//go:embed web.css
var CssFile []byte

//go:embed js.js
var JsFile []byte

//go:embed cast.js
var JsFileCast []byte

//go:embed socket.js
var JsFileSocket []byte

//go:embed dom-ready.js
var JsFileDOMReady []byte

func init() {
	WebTemplate = append(WebTemplate, []byte("\n"+`{{ define "css" }}`+"\n")...)
	WebTemplate = append(WebTemplate, CssFile...)
	WebTemplate = append(WebTemplate, []byte(`{{ end }}`+"\n")...)
	WebTemplate = append(WebTemplate, []byte("\n"+`{{ define "js" }}`+"\n")...)
	WebTemplate = append(WebTemplate, JsFile...)
	WebTemplate = append(WebTemplate, JsFileCast...)
	WebTemplate = append(WebTemplate, JsFileSocket...)
	WebTemplate = append(WebTemplate, JsFileDOMReady...)
	WebTemplate = append(WebTemplate, []byte(`{{ end }}`+"\n")...)
}
