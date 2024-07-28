package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/oktalz/present-poc/data"
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

		userID := cookieIDValue(w, r)
		isAdmin := (adminPwd == "") || cookieAdminAuth(adminPwd, r)
		// register with server
		id, serverEvent, err := server.Register(userID, isAdmin) //nolint:varnamelen
		if err != nil {
			log.Println("register:", err)
			return
		}
		defer server.Unregister(id)
		strID := id.String()
		browserEvent := make(chan data.Message)
		msg := data.Message{
			ID:     strID,
			Author: "SERVER",
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
					Author: strID,
					Msg:    message,
					Slide:  msg.Slide,
					Pool:   msg.Pool,
					Value:  msg.Value,
				}
				if msg.Pool == "" && msg.Data == nil {
					atomic.StoreInt64(&CurrentSlide, int64(msg.Slide))
				}
			}
		}(ctx)

		for {
			select {
			case msg := <-serverEvent:
				if strID == msg.Author {
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
				if msg.Pool != "" {
					// log.Println("user", userID, msg.Pool, msg.Value)
					msg.Author = userID
					server.Pool(msg)
					continue
				}

				body := data.Message{
					Author: msg.Author,
					Slide:  msg.Slide,
					Reload: false,
				}

				if isAdmin {
					// body.Author = "SERVER"
					server.Broadcast(body)
				}
			case <-ctx.Done():
				return
			}
		}
	})
}
