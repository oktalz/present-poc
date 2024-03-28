package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/oklog/ulid/v2"
	"gitlab.com/fer-go/present/data"
	"nhooyr.io/websocket"
)

var CurrentSlide = int64(-10)

func WS(server data.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println("accept:", err)
			return
		}
		defer c.Close(websocket.StatusAbnormalClosure, "error? ")

		// register with server
		id, serverEvent := server.Register()
		defer server.Unregister(id)
		browserEvent := make(chan data.Message)
		msg := data.Message{
			ID:     id,
			Author: ulid.ULID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			//Slides: data.Presentation(),
			Slide: int(atomic.LoadInt64(&CurrentSlide)),
		}

		buf, _ := json.Marshal(msg)
		err = c.Write(context.Background(), websocket.MessageText, buf)
		if err != nil {
			log.Println("write:", err)
			return
		}

		ctx := context.Background()
		go func(ctx context.Context) {
			defer ctx.Done()
			for {
				_, message, err := c.Read(context.Background())
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
					Msg:    message,
					Slide:  msg.Slide,
				}
				atomic.StoreInt64(&CurrentSlide, int64(msg.Slide))
			}
		}(ctx)

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
				err = c.Write(context.Background(), websocket.MessageText, buf)
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
				server.Broadcast(body)
			case <-ctx.Done():
				return
			}
		}
	})
}
