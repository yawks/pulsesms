package pulsesms

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type NotificationMessage struct {
	Operation string  `json:"operation,omitempty"`
	Content   Message `json:"content,omitempty"`
}

type WSMessage struct {
	Identifier string              `json:"identifier,omitempty"`
	Message    NotificationMessage `json:"message,omitempty"`
}

type Conn struct {
	conn *websocket.Conn
}

func (c *Client) Stream() {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	url := fmt.Sprintf("wss://api.pulsesms.app/api/v1/stream?account_id=%s", c.accountID)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	subscribe := map[string]interface{}{
		"command":    "subscribe",
		"identifier": "{\"channel\":\"NotificationsChannel\"}",
	}

	err = conn.WriteJSON(subscribe)
	if err != nil {
		log.Println("write:", err)
		return
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				return
			}
			c.handleMessage(message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			fmt.Println("done")
			return
		case <-interrupt:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func (c *Client) handleMessage(msg []byte) {
	wm := &WSMessage{}
	err := json.Unmarshal(msg, wm)
	if err != nil {
		fmt.Println(string(msg))
		return
	}

	if wm.Message.Operation == "" {
		return
	}

	// fmt.Println("operation:", wm.Message.Operation)
	switch wm.Message.Operation {
	case "added_message":
		m := wm.Message.Content
		err := decryptMessage(c.crypto.cipher, &m)
		if err != nil {
			fmt.Println("failed to decrypt message:", err)
			return
		}
		// update store
		convo, err := c.getConversation(m.ConversationID)
		if err != nil {
			fmt.Println(err)
			return
		}
		c.Store.SetConversation(convo)
		go c.messageHandler(m)

	case "removed_message":
	case "read_conversation":
	case "updated_conversation":
	}

}
