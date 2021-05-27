package main

import (
	"fmt"

	"github.com/treethought/pulsesms"
)


const (
	username = "me@example.com"
	password = "password"
)

func main() {
	c := pulsesms.New()
	err := c.Login(username, password)
	if err != nil {
		fmt.Println(err)
	}

	c.SetMessageHandler(func(m pulsesms.Message) {
        fmt.Printf("processing msg %v: %s", m.ID, m.Data)
        fmt.Println("getting convo msgs:", m.ConversationID)
        msgs, err := c.GetMessages(m.ConversationID, 0)
        if err != nil {
            panic(err)
        }
        fmt.Println(msgs)
	})

	convos, err := c.ListConversations()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("got %v convos\n", len(convos))

	fmt.Println("streaming")
	c.Stream()
}
