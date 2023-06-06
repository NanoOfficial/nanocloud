package common

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Packet struct {
	PType string `json:"type"`
	Data  string `json:"data"`
}

type PacketClient struct {
	conn *websocket.Conn
}

func NewPacketClient(conn *websocket.Conn) *PacketClient {
	return &PacketClient{
		conn: conn,
	}
}

func (c *PacketClient) Send(packet Packet) error {
	data, err := json.Marshal(packet)
	if err != nil {
		return err
	}

	c.conn.WriteMessage(websocket.TextMessage, data)
	return nil
}
