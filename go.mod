//module github.com/oktalz/present
module gitlab.com/fer-go/present

go 1.22

require (
	github.com/fsnotify/fsnotify v1.7.0
	github.com/gorilla/websocket v1.5.1
)

require (
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
)

replace (
	github.com/oktalz/present => gitlab.com/fer-go/present
)
