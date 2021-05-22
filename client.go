package pulsesms

import (
	"fmt"
	"path/filepath"
	"time"

	"crypto/cipher"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	api        *resty.Client
	accountID  string
	baseUrl    string
	apiVersion string
	crypto      accountCrypto
}

type accountCrypto struct {
	// the primary salt used to created thr AES keu
	salt1 []byte

	// the secondary salt used to generate the password hash
	salt2 []byte

	// has of the key derived from password and salt2
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

	return client
}

func (c *Client) getAccountParam() string {
	return "?account_id=" + c.accountID
}

func (c *Client) getUrl(endpoint string) string {
	protocol := "https://"
	if endpoint == "websocket" {
		protocol = "wss://"
	}
	url := filepath.Join(c.baseUrl, c.apiVersion, endpoint)
	fmt.Println(url)

	m := protocol + url
	fmt.Println(m)

	return protocol + url

}
