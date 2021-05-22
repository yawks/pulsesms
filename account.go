package pulsesms

import (
	"crypto/aes"
	"fmt"
)

type loginResponse struct {
	AccountID   string `json:"account_id,omitempty"`
	Salt1       string `json:"salt1,omitempty"`
	Salt2       string `json:"salt2,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Name        string `json:"name,omitempty"`
	Passcode    string `json:"passcode,omitempty"`
}

func (c *Client) Login(username, password string) error {
	body := map[string]string{
		"username": username,
		"password": password,
	}
	result := loginResponse{}

	endpoint := c.getUrl(EndpointLogin)

	resp, err := c.api.R().SetBody(body).SetResult(&result).EnableTrace().Post(endpoint)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf(resp.Status())
	}

	if result.AccountID == "" {
		return fmt.Errorf("response missing accounntID")
	}

	c.accountID = result.AccountID
	c.crypto.salt1 = []byte(result.Salt1)
	c.crypto.salt2 = []byte(result.Salt2)

	// use salt2 to generate the hash
	hash := hashPasswordSalt(password, c.crypto.salt2)
	c.crypto.pwKeyHash = hash

	// use salt1 to generate the encryption key
	c.crypto.aesKey = genAesKey(c.accountID, c.crypto.pwKeyHash, c.crypto.salt1)
	c.crypto.cipher, err = aes.NewCipher(c.crypto.aesKey)

	if err != nil {
		return err
	}

	return nil

}
