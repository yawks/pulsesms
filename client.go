package pulsesms

import (
	"fmt"
	"path/filepath"
	"time"

	"crypto/cipher"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	api            *resty.Client
	accountID      AccountID
	baseUrl        string
	apiVersion     string
	crypto         accountCrypto
	messageHandler func(Message)
	Store          *Store
	connected      bool
}

type accountCrypto struct {
	// the primary salt used to created the AES key
	salt1 []byte

	// hash of the key derived from password and salt2
	pwKeyHash string

	// the AES encryption key
	aesKey []byte

	// the AES cipher block
	cipher cipher.Block
}

func New() *Client {
	client := &Client{
		baseUrl:    "api.pulsesms.app/api",
		apiVersion: "v1",
	}

	api := resty.New()
	api.SetTimeout(60 * time.Second)
	api.SetHeaders(map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
		// "User-Agent":   clientName,
	})
	client.api = api

	client.Store = newStore()

	return client
}

func (c *Client) AccountID() AccountID {
	return c.accountID
}
func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) Sync() error {
	convos, err := c.listConversations()
	if err != nil {
		return err
	}
	for _, convo := range convos {
		chat := newChat(convo)
		c.Store.setChat(chat)
	}
	return nil
}

func (c *Client) GetChat(convoID ConversationID) (Chat, bool) {
	chat, ok := c.Store.getChatByConversationID(convoID)
	if ok {
		return chat, true
	}
	conv, err := c.getConversation(convoID)
	if err != nil {
		return Chat{}, false
	}
	chat = newChat(conv)
	c.Store.setChat(chat)
	return chat, true
}

func (c *Client) GetContactByName(name string) (Contact, bool) {
	return c.Store.getContactByName(name)
}

func (c *Client) GetContactByPhone(phone PID) (Contact, bool) {
	return c.Store.getContactByPhone(phone)
}

func (c *Client) SetMessageHandler(f func(Message)) {
	c.messageHandler = f
}

func (c *Client) getAccountParam() string {
	return fmt.Sprintf("?account_id=%s", c.accountID)
}

func (c *Client) getUrl(endpoint string) string {
	protocol := "https://"
	if endpoint == "websocket" {
		protocol = "wss://"
	}
	url := filepath.Join(c.baseUrl, c.apiVersion, endpoint)
	fmt.Println(url)
	return protocol + url

}
