package pulsesms

import (
	"crypto/aes"
	"fmt"
)

// AccountID is a PulseSMS account ID
// this reflects the Pulse SMS subscriber, not a contact or "sms user"
type AccountID string

type loginResponse struct {
	AccountID   AccountID `json:"account_id,omitempty"`
	Salt1       string    `json:"salt1,omitempty"`
	Salt2       string    `json:"salt2,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Name        string    `json:"name,omitempty"`
	Passcode    string    `json:"passcode,omitempty"`
}

// LoginCredentials is used for basic username/password login
// These are required to obtain KeyCredentials required for encryption
type BasicCredentials struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// KeyCredentials are the inputs used to generate an encryption key
// Note these can only be generated after calling Login
type KeyCredentials struct {
	// the account id required to access API resources
	AccountID AccountID

	// hash of account password and pepper (salt2)
	PasswordHash string

	// salt used
	Salt string
}

// GenerateKey generates and configures the client's encryption key
func (c *Client) GenerateKey(creds KeyCredentials) error {
	c.accountID = AccountID(creds.AccountID)
	c.crypto.salt1 = []byte(creds.Salt)
	c.crypto.pwKeyHash = creds.PasswordHash

	var err error

	c.crypto.aesKey = genAesKey(string(c.accountID), c.crypto.pwKeyHash, c.crypto.salt1)
	c.crypto.cipher, err = aes.NewCipher(c.crypto.aesKey)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetKeyCredentials() KeyCredentials {
	return KeyCredentials{
		AccountID:    c.accountID,
		PasswordHash: c.crypto.pwKeyHash,
		Salt:         string(c.crypto.salt1),
	}

}

// Login authenticates with pulse and setups up client encryption
func (c *Client) Login(creds BasicCredentials) error {
	result := loginResponse{}

	endpoint := c.getUrl(EndpointLogin)

	resp, err := c.api.R().SetBody(creds).SetResult(&result).EnableTrace().Post(endpoint)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf(resp.Status())
	}

	if result.AccountID == "" {
		return fmt.Errorf("response missing accounntID")
	}

	// use salt2 (pepper) to generate the hash
	hash := hashPasswordSalt(creds.Password, []byte(result.Salt2))

	// use salt1 to generate the encryption key
	keyCreds := KeyCredentials{
		AccountID:    result.AccountID,
		PasswordHash: hash,
		Salt:         result.Salt1,
	}

	err = c.GenerateKey(keyCreds)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SetKeyCredentials(accountID AccountID, password string, salt1 string, salt2 string) error {
	// use salt2 (pepper) to generate the hash
	hash := hashPasswordSalt(password, []byte(salt2))

	// use salt1 to generate the encryption key
	keyCreds := KeyCredentials{
		AccountID:    accountID,
		PasswordHash: hash,
		Salt:         salt1,
	}

	err := c.GenerateKey(keyCreds)
	if err != nil {
		return err
	}

	return nil
}
