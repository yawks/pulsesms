package pulsesms

import (
	"encoding/json"
	"fmt"
)

type Conversation struct {
	DeviceId     int    `json:"device_id,omitempty"`
	FolderId     int    `json:"folder_id,omitempty"`
	Read         bool   `json:"read,omitempty"`
	Timestamp    int    `json:"timestamp,omitempty"`
	Title        string `json:"title,omitempty"`
	Archive      bool   `json:"archive,omitempty"`
	Mute         bool   `json:"mute,omitempty"`
	PhoneNumbers string `json:"phone_numbers,omitempty"`
}

func (c *Client) List() error {
	index := "index_public_unarchived"

	endpoint := c.getUrl(EndpointConversations)

	path := fmt.Sprintf("%s/%s", endpoint, index)

	resp, err := c.api.R().
		SetQueryParam("account_id", c.accountID).
		SetQueryParam("limit", fmt.Sprint(75)).
		Get(path)

	if err != nil {
		fmt.Printf("%v: %s", resp.StatusCode(), resp.Status())
		return err
	}

	convos := []Conversation{}

	err = json.Unmarshal(resp.Body(), &convos)
	if err != nil {
		return fmt.Errorf("failed to unmarshall conversations: %v", err)
	}

	for _, conv := range convos {
		decrypted, err := decryptConversation(c.crypto.cipher, conv)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", decrypted)
	}

	return nil

}
