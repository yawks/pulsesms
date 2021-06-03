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
		return
	}

	fmt.Println("syncing")
	err = c.Sync()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Contacts")
	for _, contact := range c.Store.Contacts {
		fmt.Println(contact.Name, contact.PID)
	}

	fmt.Println("Chats")
	for _, chat := range c.Store.Chats {
		fmt.Println(chat.Name, chat.Members)
	}

	c.SetMessageHandler(func(m pulsesms.Message) {
		fmt.Printf("processing msg %v: %s", m.ID, m.Data)
		fmt.Println("getting convo msgs:", m.ConversationID)

        fmt.Println(m.ConversationID)
        fmt.Println(m.DeviceID)
        fmt.Println(m.Data)

        fmt.Println("from convo id")
        convo := c.Store.Chats[m.ConversationID]
        fmt.Println(convo)
	})

	convos, err := c.ListConversations()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("got %v convos\n", len(convos))

	fmt.Println("streaming")
	c.Stream()
}
