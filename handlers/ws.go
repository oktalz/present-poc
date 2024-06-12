package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/oklog/ulid/v2"
	"gitlab.com/fer-go/present/data"
	"gitlab.com/fer-go/present/hash"
	"nhooyr.io/websocket"
)

var CurrentSlide = int64(-10)

func WS(server data.Server, adminPwd string) http.Handler { //nolint:funlen,gocognit
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Println("accept:", err)
			return
		}
		defer conn.Close(websocket.StatusAbnormalClosure, "error? ")

		// register with server
		id, serverEvent := server.Register() //nolint:varnamelen
		defer server.Unregister(id)
		browserEvent := make(chan data.Message)
		msg := data.Message{
			ID:     id,
			Author: ulid.ULID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			// Slides: data.Presentation(),
			Slide: int(atomic.LoadInt64(&CurrentSlide)),
		}

		buf, _ := json.Marshal(msg)
		err = conn.Write(context.Background(), websocket.MessageText, buf) //nolint:contextcheck
		if err != nil {
			log.Println("write:", err)
			return
		}

		ctx := context.Background()
		go func(ctx context.Context) { //nolint:contextcheck
			defer ctx.Done()
			for {
				_, message, err := conn.Read(context.Background()) //nolint:contextcheck
				if err != nil {
					log.Println("read:", id, err)
					return
				}
				var msg data.Message
				err = json.Unmarshal(message, &msg)
				if err != nil {
					log.Println(err)
					continue
				}
				browserEvent <- data.Message{
					Author: id,
					Admin:  msg.Admin,
					Msg:    message,
					Slide:  msg.Slide,
				}
				atomic.StoreInt64(&CurrentSlide, int64(msg.Slide))
			}
		}(ctx)

		isAdmin := false
		if adminPwd == "" {
			isAdmin = true
		} else {
			var pass string
			cookie, err := r.Cookie("present")
			if err == nil {
				// Cookie exists, you can access its value using cookie.Value
				pass = cookie.Value
				isAdmin = hash.Equal(pass, adminPwd)
			}
		}

		for {
			select {
			case msg := <-serverEvent:
				if id == msg.Author {
					continue
				}
				buf, _ := json.Marshal(msg)
				// log.Println(id, string(buf))
				if msg.Reload {
					msg.Slide = int(CurrentSlide)
				}
				err = conn.Write(context.Background(), websocket.MessageText, buf) //nolint:contextcheck
				if err != nil {
					log.Println("write:", err)
					return
				}
			case msg := <-browserEvent:
				body := data.Message{
					Author: msg.Author,
					Slide:  msg.Slide,
					Reload: false,
				}

				if msg.Admin == adminPwd {
					isAdmin = true
				}
				if isAdmin {
					server.Broadcast(body)
				}
			case <-ctx.Done():
				return
			}
		}
	})
}
