package main

import (
	_ "embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"gitlab.com/fer-go/present/archive"
	"gitlab.com/fer-go/present/data"
	"gitlab.com/fer-go/present/handlers"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	zipFlag := false
	for _, arg := range os.Args[1:] {
		if arg == "zip" {
			zipFlag = true
			break
		}
	}
	wd := ""
	if !zipFlag && len(os.Args) > 1 {
		// osarg could be also an url, in that case i need to download the targx, unzip it

		wd = os.Args[1]
		if fileInfo, err := os.Stat(wd); err == nil {
			if !fileInfo.IsDir() {
				if err := archive.UnGzip(wd); err != nil {
					panic(err)
				}
			}
		} else {
			err := os.Chdir(wd)
			if err != nil {
				panic(err)
			}
		}
	}
	if zipFlag {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		err = archive.Gzip(wd, "present.tar.gz")
		if err != nil {
			panic(err)
		}
		os.Exit(0)
		return
	}
	wsServer := data.NewServer()
	data.Init(wsServer)

	http.Handle("/api", handlers.API())
	http.Handle("/exec", handlers.Exec())
	http.Handle("/cast", handlers.CastWS())
	http.Handle("/asciinema", handlers.Asciinema())

	http.Handle("/ws", handlers.WS(wsServer))

	// Serve static files
	sub, err := fs.Sub(dist, "ui/dist")
	if err != nil {
		panic(err)
	}
	wd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	handler := &fallbackFileServer{
		primary:   http.FileServer(http.FS(sub)),
		secondary: http.FileServer(http.Dir(wd)),
	}
	// http.Handle("/", http.FileServer(http.FS(sub)))
	http.Handle("/", handler)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	go func() {
		err = http.ListenAndServe(":8081", newUI())
		if err != nil {
			panic(err)
		}
	}()

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
