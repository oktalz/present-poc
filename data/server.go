package data

import (
	"log"
	"sync"

	"github.com/oklog/ulid/v2"
)

var (
	muWS   sync.RWMutex
	poolWS sync.RWMutex
)

type Server interface {
	Register(userID string, isAdmin bool) (id ulid.ULID, ch chan Message, err error)
	Unregister(id ulid.ULID)
	Broadcast(msg Message)
	Send(id ulid.ULID, msg Message)
	Pool(msg Message)
}

func NewServer() *server { //nolint:revive
	return &server{
		clients: make(map[ulid.ULID]chan Message),
		admins:  make(map[ulid.ULID]chan Message),
		pools:   make(map[string]map[string]string),
	}
}

type server struct {
	clients map[ulid.ULID]chan Message
	admins  map[ulid.ULID]chan Message
	pools   map[string]map[string]string
}

func (s *server) Register(userID string, isAdmin bool) (id ulid.ULID, ch chan Message, err error) { //nolint:nonamedreturns
	muWS.Lock()
	defer muWS.Unlock()
	id, err = ulid.Parse(userID)
	if err != nil {
		log.Println(err)
		return ulid.ULID{}, nil, err
	}
	ch = make(chan Message)
	if isAdmin {
		s.admins[id] = ch
		log.Println("registered admin", id)
	} else {
		s.clients[id] = ch
		log.Println("registered", id)
	}
	s.clients[id] = ch
	return id, ch, nil
}

func (s *server) Unregister(id ulid.ULID) {
	muWS.Lock()
	defer muWS.Unlock()
	log.Println("unregistered", id)
	delete(s.clients, id)
	delete(s.admins, id)
}

func (s *server) Broadcast(msg Message) {
	muWS.RLock()
	defer muWS.RUnlock()
	// log.Println("broadcast", msg.Author)
	for _, ch := range s.admins {
		go func(ch chan Message, msg Message) {
			ch <- msg
		}(ch, msg)
	}
	for _, ch := range s.clients {
		go func(ch chan Message, msg Message) {
			ch <- msg
		}(ch, msg)
	}
}

func (s *server) BroadcastAdmins(msg Message) {
	muWS.RLock()
	defer muWS.RUnlock()
	for _, ch := range s.admins {
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

func (s *server) Pool(msg Message) {
	muWS.RLock()
	defer muWS.RUnlock()
	pool, ok := s.pools[msg.Pool]
	if ok {
		if pool[msg.Author] == msg.Value {
			return
		}
		pool[msg.Author] = msg.Value
	} else {
		s.pools[msg.Pool] = make(map[string]string)
		s.pools[msg.Pool][msg.Author] = msg.Value
	}
	d := map[string]int{}
	for _, v := range s.pools[msg.Pool] {
		d[v]++
	}

	bMsg := Message{
		Pool: msg.Pool,
		Data: d,
	}
	go s.Broadcast(bMsg)
}
