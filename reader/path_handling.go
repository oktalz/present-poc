package reader

import (
	"fmt"
	"os"
	"path/filepath"
)

func getOSPath(path string) string {
	path = filepath.FromSlash(path)
	fileInfo, err := os.Stat(path)
	if err != nil {
		// maybe we are on the different platform
		path = convertToOSPath(path)
		fileInfo, err = os.Stat(path)
		if err != nil {
			// fmt.Println(err.Error())
			return ""
		}
	}

	if fileInfo.IsDir() {
		return path
	} else {
		fmt.Println(path, "is not a directory")
	}
	return path
}
