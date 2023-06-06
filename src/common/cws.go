package common

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	id               string
	conn             *websocket.Conn
	sendLock         sync.Mutex
	sendCallback     map[string]func(req WSPacket)
	sendCallbackLock sync.Mutex
	recvCallback     map[string]func(req WSPacket)
	Done             chan struct{}
}

type WSPacket struct {
	Type      string `json:"type"`
	Data      string `json:"data"`
	PacketID  string `json:"packet_id"`
	SessionID string `json:"session_id"`
}

var EmptyPacket = WSPacket{}

func NewClient(conn *websocket.Conn) *Client {
	id := uuid.Must(uuid.NewV4()).String()
	sendCallback := map[string]func(WSPacket){}
	recvCallback := map[string]func(WSPacket){}

	return &Client{
		id:   id,
		conn: conn,

		sendCallback: sendCallback,
		recvCallback: recvCallback,

		Done: make(chan struct{}),
	}
}

func (c *Client) Send(request WSPacket, callback func(response WSPacket)) {
	request.PacketID = uuid.Must(uuid.NewV4()).String()
	data, err := json.Marshal(request)
	if err != nil {
		return
	}

	if callback != nil {
		wrapperCallback := func(resp WSPacket) {
			defer func() {
				if err := recover(); err != nil {
					log.Println("Recovered from err in client callback ", err)
				}
			}()

			resp.PacketID = request.PacketID
			resp.SessionID = request.SessionID
			callback(resp)
		}
		c.sendCallbackLock.Lock()
		c.sendCallback[request.PacketID] = wrapperCallback
		c.sendCallbackLock.Unlock()
	}

	c.sendLock.Lock()
	c.conn.SetWriteDeadline(time.Now().Add(20 * time.Second))
	c.conn.WriteMessage(websocket.TextMessage, data)
	c.sendLock.Unlock()
}

func (c *Client) Receive(id string, f func(request WSPacket) (response WSPacket)) {
	c.recvCallback[id] = func(request WSPacket) {
		resp := f(request)
		resp.PacketID = request.PacketID
		resp.SessionID = request.SessionID

		if resp == EmptyPacket {
			return
		}
		respText, err := json.Marshal(resp)
		if err != nil {
			log.Println("[!] json marshal error:", err)
		}
		c.sendLock.Lock()
		c.conn.SetWriteDeadline(time.Now().Add(20 * time.Second))
		c.conn.WriteMessage(websocket.TextMessage, respText)
		c.sendLock.Unlock()
	}
}

func (c *Client) SyncSend(request WSPacket) (response WSPacket) {
	res := make(chan WSPacket)
	f := func(resp WSPacket) {
		res <- resp
	}
	c.Send(request, f)
	return <-res
}

func (c *Client) Heartbeat() {
	timer := time.Tick(time.Second)

	for range timer {
		select {
		case <-c.Done:
			log.Println("Close heartbeat")
			return
		default:
		}
		c.Send(WSPacket{Type: "heartbeat"}, nil)
	}
}

func (c *Client) Listen() {
	for {
		c.conn.SetReadDeadline(time.Now().Add(20 * time.Second))
		_, rawMsg, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("[!] read:", err)
			close(c.Done)
			break
		}
		wspacket := WSPacket{}
		err = json.Unmarshal(rawMsg, &wspacket)

		if err != nil {
			log.Println("Warn: error decoding", rawMsg)
			continue
		}

		callback, ok := c.sendCallback[wspacket.PacketID]

		if ok {
			go callback(wspacket)
			delete(c.sendCallback, wspacket.PacketID)
			continue
		}

		if callback, ok := c.recvCallback[wspacket.Type]; ok {
			go callback(wspacket)
		}
	}
}

func (c *Client) GetID() string {
	return c.id
}

func (c *Client) Close() {
	if c == nil || c.conn == nil {
		return
	}
	c.conn.Close()
}
