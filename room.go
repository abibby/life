package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Life int    `json:"life"`
}

type Room struct {
	Code        string    `json:"code"`
	Players     []*Player `json:"players"`
	CurrentTurn int       `json:"current_turn"`

	mtx     *sync.RWMutex
	clients []*websocket.Conn
}

type WSRequest struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type WSError struct {
	Message string `json:"message"`
}

var roomsMtx = &sync.RWMutex{}
var rooms = map[string]*Room{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	vars := mux.Vars(r)

	roomID := vars["room"]
	playerID := vars["player"]
	room, err := getRoom(ws, roomID, playerID)
	if err != nil {
		log.Print(err)
		ws.Close()
		return
	}

	defer func() {
		room.mtx.Lock()
		defer room.mtx.Unlock()

		for i, c := range room.clients {
			if c == ws {
				room.clients[i] = room.clients[len(room.clients)-1]
				room.clients = room.clients[:len(room.clients)-1]
				return
			}
		}
	}()

	room.mtx.RLock()
	err = ws.WriteJSON(room)
	room.mtx.RUnlock()
	if err != nil {
		log.Print(err)
	}

	req := &WSRequest{}
	for {
		err = ws.ReadJSON(req)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		switch req.Type {
		case "change":
			change(room, playerID, req)
		case "set-name":
			setName(room, playerID, req)
		}

		room.mtx.RLock()
		clients := room.clients
		room.mtx.RUnlock()
		for _, c := range clients {
			err = c.WriteJSON(room)
			if err != nil {
				log.Print(err)
				continue
			}
		}
	}
}

func getRoom(ws *websocket.Conn, roomID, playerID string) (*Room, error) {
	player := &Player{
		ID:   playerID,
		Name: "",
		Life: 20,
	}
	roomsMtx.RLock()
	room, ok := rooms[roomID]
	roomsMtx.RUnlock()
	if !ok {
		room = &Room{
			Code: roomID,
			Players: []*Player{
				player,
			},
			mtx: &sync.RWMutex{},
		}
		roomsMtx.Lock()
		rooms[roomID] = room
		roomsMtx.Unlock()
	}

	hasPlayer := false
	room.mtx.RLock()
	for _, p := range room.Players {
		if p.ID == playerID {
			hasPlayer = true
		}
	}
	room.mtx.RUnlock()

	room.mtx.Lock()
	if !hasPlayer {
		room.Players = append(room.Players, player)
	}
	room.clients = append(room.clients, ws)
	room.mtx.Unlock()

	return room, nil
}

type ChangeData int

func change(room *Room, playerID string, r *WSRequest) {
	c := ChangeData(0)
	err := json.Unmarshal(r.Data, &c)
	if err != nil {
		log.Print(err)
		return
	}
	room.mtx.Lock()
	defer room.mtx.Unlock()

	for _, p := range room.Players {
		if p.ID == playerID {
			p.Life += int(c)
		}
	}
}

type SetNameData string

func setName(room *Room, playerID string, r *WSRequest) {
	name := SetNameData("")
	err := json.Unmarshal(r.Data, &name)
	if err != nil {
		log.Print(err)
		return
	}
	room.mtx.Lock()
	defer room.mtx.Unlock()

	for _, p := range room.Players {
		if p.ID == playerID {
			p.Name = string(name)
		}
	}
}
