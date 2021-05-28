package pulsesms

import (
	"encoding/json"
	"fmt"
)

// ConversationID is the internal ID of a group or one-on-one chat / thread
type ConversationID = int


type Conversation struct {
	ID           ConversationID `json:"id,omitempty"`
	DeviceId     int            `json:"device_id,omitempty"`
	FolderId     int            `json:"folder_id,omitempty"`
	Read         bool           `json:"read,omitempty"`
	Timestamp    int64          `json:"timestamp,omitempty"`
	Title        string         `json:"title,omitempty"`
	Archive      bool           `json:"archive,omitempty"`
	Mute         bool           `json:"mute,omitempty"`
	PhoneNumbers string         `json:"phone_numbers,omitempty"`
	Snippet      string         `json:"snippet,omitempty"`
}

func (c *Client) ListConversations() ([]Conversation, error) {
	index := "index_public_unarchived"

	endpoint := c.getUrl(EndpointConversations)

	path := fmt.Sprintf("%s/%s", endpoint, index)

	resp, err := c.api.R().
		SetQueryParam("account_id", fmt.Sprint(c.accountID)).
		SetQueryParam("limit", fmt.Sprint(75)).
		Get(path)

	if err != nil {
		fmt.Printf("%v: %s", resp.StatusCode(), resp.Status())
		return nil, err

	}

	convos := []Conversation{}

	err = json.Unmarshal(resp.Body(), &convos)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall conversations: %v", err)
	}

	result := []Conversation{}
	for _, conv := range convos {
		err := decryptConversation(c.crypto.cipher, &conv)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt conversation %v", err)
		}
		result = append(result, conv)

	}

	return result, nil

}

func (c *Client) updateConversation(conversationID int, snippet string, timestamp int64) error {
	req := updateConversationRequest{
		AccountID: c.accountID,
		Read:      false,
		Timestamp: timestamp,
		Snippet:   snippet,
	}

	endpoint := c.getUrl(EndpointUpdateConversation)
	endpoint = fmt.Sprintf("%s/%s", endpoint, fmt.Sprint(conversationID))
	resp, err := c.api.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(endpoint)

	if resp.StatusCode() > 200 || err != nil {
		fmt.Println(endpoint)
		fmt.Println(resp.StatusCode(), resp.Status())
		fmt.Printf(string(resp.Body()))
		return err
	}
	fmt.Println("updated conversation", conversationID)
	return nil
}
