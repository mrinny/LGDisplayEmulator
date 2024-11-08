package webapp

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 1 * time.Second
	pongWait       = 10 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	id   string
	conn *websocket.Conn
	send chan []byte
	hub  *Hub
}

func (c *Client) writePump() {
	slog.Info("starting writePump")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		slog.Info("write pump closed", "id", c.id)
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			n, err := w.Write(msg)
			if err != nil {
				slog.Error("error while writing to client", "id", c.id)
				return
			}
			slog.Debug("written bytes to client", "client", c.id, "bytes", n)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	slog.Info("starting readPump")
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, text, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("erreur", "error", err)
			}
			break
		}

		msg := &ActionRequest{}

		reader := bytes.NewReader(text)
		decoder := json.NewDecoder(reader)
		err = decoder.Decode(msg)
		if err != nil {
			slog.Error("erreur", "error", err)
		}
		c.hub.actions <- msg
	}
}

type ActionRequest struct {
	Action string `json:"action"`
	Id     int    `json:"id"`
	Input  string `json:"input"`
}
