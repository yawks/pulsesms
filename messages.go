package pulsesms

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

type Message struct {
	ID             string `json:"id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	DeviceID       string `json:"device_id,omitempty"`
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

func (c *Client) GetMessages(conversationID string, offset int) {
	const limit = 70

	endpoint := c.getUrl(EndpointMessages)

	resp, err := c.api.R().
		SetQueryParam("account_id", c.accountID).
		SetQueryParam("conversation_id", conversationID).
		SetQueryParam("offset", fmt.Sprint(offset)).
		SetQueryParam("limit", fmt.Sprint(limit)).
		Get(endpoint)

	data, err := ioutil.ReadAll(bytes.NewReader(resp.Body()))
	if err != nil {
		panic(err)
	}

	fmt.Printf(string(data))

}
