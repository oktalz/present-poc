package data

import (
	"log"
	"sync"

	"github.com/oklog/ulid/v2"
)

var muWS sync.RWMutex

type Server interface {
	Register() (id ulid.ULID, ch chan Message)
	Unregister(ulid.ULID)
	Broadcast(msg Message)
	Send(id ulid.ULID, msg Message)
}

func NewServer() *server {
	return &server{
		clients: make(map[ulid.ULID]chan Message),
	}
}

type server struct {
	clients map[ulid.ULID]chan Message
}

func (s *server) Register() (id ulid.ULID, ch chan Message) {
	muWS.Lock()
	defer muWS.Unlock()
	id = ulid.Make()
	log.Println("registered", id)
	ch = make(chan Message)
	s.clients[id] = ch
	return id, ch
}

func (s *server) Unregister(id ulid.ULID) {
	muWS.Lock()
	defer muWS.Unlock()
	log.Println("unregistered", id)
	delete(s.clients, id)
}

func (s *server) Broadcast(msg Message) {
	muWS.RLock()
	defer muWS.RUnlock()
	// log.Println("broadcast", msg.Author)
	for _, ch := range s.clients {
		go func(ch chan Message, msg Message) {
			ch <- msg
		}(ch, msg)
	}
}

func (s *server) Send(id ulid.ULID, msg Message) {
	muWS.RLock()
	defer muWS.RUnlock()
	ch, ok := s.clients[id]
	if ok {
		ch <- msg
	}
}
