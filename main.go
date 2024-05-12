package main

import (
	"embed"
	"io/fs"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"

	"gitlab.com/fer-go/present/archive"
	"gitlab.com/fer-go/present/data"
	"gitlab.com/fer-go/present/handlers"
)

func RandomString() string {
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	var sb []rune
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 12; i++ {
		sb = append(sb, chars[rand.Intn(len(chars))])
	}
	return string(sb)
}

func main() { //nolint:funlen
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stderr)

	_ = godotenv.Load()
	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	var portInt int
	var err error
	if portInt, err = strconv.Atoi(port); err != nil {
		panic(err)
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	zipFlag := false
	for _, arg := range os.Args[1:] {
		if arg == "zip" {
			zipFlag = true
			break
		}
	}
	var wd string                     //nolint:varnamelen
	if !zipFlag && len(os.Args) > 1 { //nolint:nestif
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

	// http.Handle("/api", handlers.API())
	// http.Handle("/exec", handlers.Exec())
	adminPWD := os.Getenv("ADMIN")
	if adminPWD != "" {
		log.Println("admin token is set")
		log.Println("☢☢☢")
		log.Println("set localStorage.setItem('admin-token', '') in browser !!!")
		log.Println("set correct token !!! 🙈🙉🙊")
	}

	// http.Handle("/cast", handlers.Cast())
	http.Handle("/cast", handlers.CastWS(wsServer, adminPWD))
	http.Handle("/asciinema", handlers.Asciinema())
	http.Handle("/ws", handlers.WS(wsServer, adminPWD))

	http.Handle("/{$}", handlers.Homepage(portInt))

	sub, err := fs.Sub(dist, "ui/static")
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
	http.Handle("/", handler)
	// http.Handle("/", http.FileServer(http.Dir(wd)))
	server := &http.Server{
		//		Addr:         ":" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// log.Println("Listening on", server.Addr)
	// err = server.ListenAndServe()
	// if err != nil {
	// 	panic(err)
	// }

	var errList []error
	wg := new(sync.WaitGroup)
	// Create a TCP listener for IPv4
	ln4, err := net.Listen("tcp4", ":"+port)
	if err == nil {
		wg.Add(1)
		go func() {
			if err := server.Serve(ln4); err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}()
	} else {
		errList = append(errList, err)
	}

	// Create a TCP listener for IPv6
	ln6, err := net.Listen("tcp6", "[::]:"+port)
	if err == nil {
		wg.Add(1)
		go func() {
			if err := server.Serve(ln6); err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}()
	} else {
		errList = append(errList, err)
	}
	if len(errList) > 1 {
		panic(errList)
	}

	wg.Wait()
}

//go:embed ui/static
var dist embed.FS
