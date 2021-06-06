package main

import (
	"fmt"

	"github.com/treethought/pulsesms"
)

var creds = pulsesms.BasicCredentials{
	Username: "me@example.com",
	Password: "passwoord",
}

func main() {
	c := pulsesms.New()

	err := c.Login(creds)
	if err != nil {
		fmt.Println(err)
		return
	}

	// or if you have already logged in
	// and need reconfigure encryption
	keyCreds := c.GetKeyCredentials()
	err = c.GenerateKey(keyCreds)
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

	fmt.Println("\nContacts")
	for _, contact := range c.Store.Contacts {
		fmt.Println(contact.Name)
		fmt.Println(contact.PID)
		fmt.Println("")
	}

	fmt.Println("\nChats")
	for _, chat := range c.Store.Chats {
		fmt.Println(chat.Name)
		fmt.Println(chat.ID)
		fmt.Println("")
	}

	c.SetMessageHandler(func(m pulsesms.Message) {

		fmt.Println(m.ConversationID)
		fmt.Println("getting conversation")
		chat, ok := c.GetChat(m.ChatID())
		if !ok {
			fmt.Println("couldnt find convo")
		}
		fmt.Println("message from chat:", chat.ID)
		fmt.Println("members", chat.Members)

	})

	fmt.Println("streaming")
	c.Stream()
}
