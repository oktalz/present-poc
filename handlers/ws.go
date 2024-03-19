package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"
	"gitlab.com/fer-go/present/data"
)

var upgraderWS = websocket.Upgrader{
	// Allow CORS
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by always returning true
		return true
	},
}

func WS(server data.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		// register with server
		id, serverEvent := server.Register()
		log.Println(id)
		defer server.Unregister(id)
		browserEvent := make(chan data.Message)
		msg := data.Message{
			ID:     id,
			Author: ulid.ULID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			//Slides: data.Presentation(),
			Slide: -10,
		}

		buf, _ := json.Marshal(msg)
		err = c.WriteMessage(websocket.TextMessage, buf)
		if err != nil {
			log.Println("write:", err)
			return
		}

		ctx := context.Background()
		go func(ctx context.Context) {
			defer ctx.Done()
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("error: %v", err)
					}
					break
				}
				var msg data.Message
				err = json.Unmarshal(message, &msg)
				if err != nil {
					log.Println(err)
					continue
				}
				if msg.Author != id {
					continue
				}
				browserEvent <- data.Message{
					Author: id,
					Msg:    message,
					Slide:  msg.Slide,
				}
			}
		}(ctx)

		for {
			select {
			case msg := <-serverEvent:
				if msg.ID == id && msg.ID == msg.Author {
					continue
				}
				buf, _ := json.Marshal(msg)
				log.Println(string(buf))
				err = c.WriteMessage(websocket.TextMessage, buf)
				if err != nil {
					log.Println("write:", err)
					return
				}
			case msg := <-browserEvent:
				body := data.Message{
					ID:     id,
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
