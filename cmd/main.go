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

    convos, err := c.List()
	if err != nil {
		fmt.Println(err)
	}
    fmt.Println(convos)

    fmt.Println("streaming")
	c.Stream()
}
