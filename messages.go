package pulsesms

import (
	"fmt"
)

type Message struct {
	ID             int    `json:"id,omitempty"`
	ConversationID int    `json:"conversation_id,omitempty"`
	DeviceID       int    `json:"device_id,omitempty"`
	Type           int    `json:"type,omitempty"`
	Data           string `json:"data,omitempty"`
	Timestamp      int    `json:"timestamp,omitempty"`
	MimeType       string `json:"mime_type,omitempty"`
	Read           bool   `json:"read,omitempty"`
	Seen           bool   `json:"seen,omitempty"`
	From           string `json:"from,omitempty"`
	Archive        bool   `json:"archive,omitempty"`
	SentDevice     int    `json:"sent_device,omitempty"`
	SimStamp       string `json:"sim_stamp,omitempty"`
	Snippet        string `json:"snippet,omitempty"`
}

func (c *Client) GetMessages(conversationID int, offset int) ([]Message, error) {
	msgs := []Message{}
	const limit = 70

	endpoint := c.getUrl(EndpointMessages)

	resp, err := c.api.R().
		SetQueryParam("account_id", c.accountID).
		SetQueryParam("conversation_id", fmt.Sprint(conversationID)).
		SetQueryParam("offset", fmt.Sprint(offset)).
		SetQueryParam("limit", fmt.Sprint(limit)).
		SetResult(&msgs).
		Get(endpoint)

	if err != nil {
		fmt.Printf("%v: %s\n", resp.StatusCode(), resp.Status())
		return nil, err
	}

	result := []Message{}
	for _, m := range msgs {
		err := decryptMessage(c.crypto.cipher, &m)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt message %v", err)
		}
		result = append(result, m)

	}

	return result, nil

}
