//go:build windows

package reader

import "strings"

const TerminalChar = ">"

func convertToOSPath(path string) string {
	return strings.ReplaceAll(path, "/", "\\")
}
