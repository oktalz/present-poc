package main

import (
	"embed"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/oktalz/present-poc/archive"
	"github.com/oktalz/present-poc/data"
	"github.com/oktalz/present-poc/handlers"
)

//go:embed ui/static
var dist embed.FS

//go:embed ui/login.html
var loginPage []byte

func main() { //nolint:funlen
	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer f.Close()
	// _ = pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// f, err := os.Create("mem.prof")
	// if err != nil {
	// 	log.Fatal("could not create memory profile: ", err)
	// }
	// defer f.Close()
	// defer pprof.WriteHeapProfile(f)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stderr) // why do packages feel the need to change this in init()?

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
	adminPWD := os.Getenv("ADMIN_PWD")
	if adminPWD != "" {
		log.Println("admin token is set ☢☢☢ 🙈🙉🙊")
		// adminPWD, err = hash.Hash(adminPWD)
		// if err != nil {
		// 	panic(err)
		// }
	}
	userPwd := os.Getenv("USER_PWD")
	if userPwd != "" {
		log.Println("user password is set")
		// userPwd, err = hash.Hash(userPwd)
		// if err != nil {
		// 	panic(err)
		// }
	}

	// http.Handle("/cast", handlers.Cast())
	http.Handle("/cast", handlers.CastWS(wsServer, adminPWD))
	http.Handle("/asciinema", handlers.Asciinema())
	http.Handle("/ws", handlers.WS(wsServer, adminPWD))

	http.Handle("/{$}", handlers.Homepage(portInt, userPwd, adminPWD))
	http.Handle("/login", handlers.Login(loginPage))
	http.Handle("/api/login", handlers.APILogin(userPwd, adminPWD))

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
