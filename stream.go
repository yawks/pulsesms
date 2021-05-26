package pulsesms

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) Stream() {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	url := "wss://api.pulsesms.app/api/v1/stream?account_id=" + c.accountID
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println(resp.Status)
	fmt.Println(resp.StatusCode)
	for k, v := range resp.Header {
		fmt.Println(k, v)
	}

	for k, v := range resp.Request.Header {
		fmt.Println(k, v)
	}

	conn.SetPingHandler(func(d string) error {
		fmt.Println("ping")
		return conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(time.Minute*1))

	})

	fmt.Println("subscribing")
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
			fmt.Println("reading message")
			ty, message, err := conn.ReadMessage()
			fmt.Println("type", ty)
			if err != nil {
				fmt.Println("read:", err)
				return
			}
			fmt.Printf("recv: %s", message)
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
