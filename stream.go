package pulsesms

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
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

type MessageAction int64

const (
	ReceivedMessage       MessageAction = 0
	SentMessage                         = 1
	ReadConversation                    = 2
	RemovedMessage                      = 3
	UpdatedConversation                 = 4
	DismissedNotification               = 5
	ConnectionError                     = 999
)

func (c *Client) Disconnect() {
	c.connected = false
	c.conn.Close()
	c.conn = nil
}

func (c *Client) Stream() error {
	if c.conn != nil {
		c.Disconnect()
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	url := fmt.Sprintf("wss://api.pulsesms.app/api/v1/stream?account_id=%s", c.accountID)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	c.conn = conn
	c.connected = true
	defer c.Disconnect()

	subscribe := map[string]interface{}{
		"command":    "subscribe",
		"identifier": "{\"channel\":\"NotificationsChannel\"}",
	}

	err = c.conn.WriteJSON(subscribe)
	if err != nil {
		return err
	}

	done := make(chan struct{})
	isWSOpen := true
	go func() {
		defer close(done)
		for isWSOpen {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				isWSOpen = false
				fmt.Println("read:", err)
				m := Message{}
				m.ID = -1
				m.ConversationID = -1
				go c.messageHandler(m, ConnectionError)
				break
			}
			c.handleMessage(message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case <-interrupt:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				fmt.Println("write close:", err)
				return nil
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}

func (c *Client) handleMessage(msg []byte) {
	rId := regexp.MustCompile("\"(id|android_device|device_id|message_type)\":\"(\\d+)\"")
	finalMsg := rId.ReplaceAllString(string(msg), "\"$1\":$2")

	wm := &WSMessage{}
	err := json.Unmarshal([]byte(finalMsg), wm)
	if err != nil {
		if !strings.Contains(string(msg), "ping") {
			fmt.Println(string(msg))
		}
		return
	}

	if !strings.Contains(string(msg), "ping") {
		err = nil
	}

	if wm.Message.Operation == "" {
		return
	}

	fmt.Println("operation:", wm.Message.Operation)
	m := wm.Message.Content
	err = decryptMessage(c.crypto.cipher, &m)
	if err != nil {
		fmt.Println("failed to decrypt message:", err)
		return
	}
	// update store
	if m.ConversationID == 0 {
		m.ConversationID = wm.Message.Content.ID
	}
	convo, err := c.getConversation(m.ConversationID)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch wm.Message.Operation {
	case "added_message":
		c.Store.setConversation(convo)
		if wm.Message.Content.Type == 0 {
			go c.messageHandler(m, ReceivedMessage)
		} else if wm.Message.Content.Type == 2 {
			go c.messageHandler(m, SentMessage)
		} else {
			fmt.Println("Unhandled added_message type")
		}
	case "read_conversation":
		go c.messageHandler(m, ReadConversation)
	case "removed_message":
	case "updated_conversation":
	case "dismissed_notification":
	}

}
