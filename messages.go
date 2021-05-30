package pulsesms

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"
)

// MessageID is the internal ID of a Pulse SMS message
type MessageID = int

// DeviceID is the generated internal ID of the device used to interact with a PulseSMS account
type DeviceID = int

type Message struct {
	ID             MessageID      `json:"id,omitempty"`
	ConversationID ConversationID `json:"conversation_id,omitempty"`
	DeviceID       DeviceID       `json:"device_id,omitempty"`
	Type           int            `json:"type,omitempty"`
	Data           string         `json:"data,omitempty"`
	Timestamp      int64          `json:"timestamp,omitempty"`
	MimeType       string         `json:"mime_type,omitempty"`
	Read           bool           `json:"read,omitempty"`
	Seen           bool           `json:"seen,omitempty"`
	From           string         `json:"message_from,omitempty"`
	Archive        bool           `json:"archive,omitempty"`
	SentDevice     DeviceID       `json:"sent_device,omitempty"`
	SimStamp       string         `json:"sim_stamp,omitempty"`
	Snippet        string         `json:"snippet,omitempty"`
}

type sendMessageRequest struct {
	AccountID            AccountID `json:"account_id,omitempty"`
	Data                 string    `json:"data,omitempty"`
	DeviceConversationID int       `json:"device_conversation_id,omitempty"`
	DeviceID             DeviceID  `json:"device_id,omitempty"`
	MessageType          int       `json:"message_type,omitempty"`
	MimeType             string    `json:"mime_type,omitempty"`
	Read                 bool      `json:"read,omitempty"`
	Seen                 bool      `json:"seen,omitempty"`
	SentDevice           DeviceID  `json:"sent_device"`
	Timestamp            int64     `json:"timestamp,omitempty"`
}

type updateConversationRequest struct {
	AccountID AccountID `json:"account_id,omitempty"`
	Read      bool      `json:"read,omitempty"`
	Timestamp int64     `json:"timestamp,omitempty"`
	Snippet   string    `json:"snippet,omitempty"`
}

func generateID() int {
	const min = 1
	const max = 922337203685477

	s := rand.Float64()
	x := s * (max - min + 1)

	return int(math.Floor(x) + min)
}

func (c *Client) GetMessages(conversationID int, offset int) ([]Message, error) {
	msgs := []Message{}
	const limit = 70

	endpoint := c.getUrl(EndpointMessages)

	resp, err := c.api.R().
		SetQueryParam("account_id", fmt.Sprint(c.accountID)).
		SetQueryParam("conversation_id", fmt.Sprint(conversationID)).
		SetQueryParam("offset", fmt.Sprint(offset)).
		SetQueryParam("limit", fmt.Sprint(limit)).
		SetResult(&msgs).
		Get(endpoint)

	if resp.StatusCode() > 200 || err != nil {
		fmt.Printf("%v: %s\n", resp.StatusCode(), resp.Status())
		return nil, err
	}
	// fmt.Println(string(resp.Body()))

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

func (c *Client) SendMessage(m Message, convoID ConversationID) error {
	cID, err := strconv.Atoi(convoID)
	if err != nil {
		return fmt.Errorf("invalid convoID: %s, %v", convoID, err)
	}

	if m.ID == 0 {
		m.ID = generateID()
	}
	if m.Snippet == "" {
		// TODO accept mimetype
		m.Snippet = fmt.Sprintf("You: %s", m.Data)
	}

	mime := "text/plain"
	if m.MimeType == "" {
		m.MimeType = mime
	}

	timestamp := time.Now().Unix()
	if m.Timestamp == 0 {
		m.Timestamp = timestamp
	}

	if m.Type == 0 {
		m.Type = 2
	}

	mimetype, err := encrypt(c.crypto.cipher, "text/plain")
	if err != nil {
		return err
	}
	encData, err := encrypt(c.crypto.cipher, m.Data)
	if err != nil {
		return err
	}
	encSnippet, err := encrypt(c.crypto.cipher, m.Snippet)
	if err != nil {
		return err
	}


	req := sendMessageRequest{
		AccountID:            c.accountID,
		Data:                 encData,
		DeviceConversationID: cID,
		// DeviceID:             id,
		DeviceID:    m.ID,
		MessageType: 2,
		Timestamp:   timestamp,
		MimeType:    mimetype,
		Read:        false,
		Seen:        false,
		SentDevice:  1,
	}

	endpoint := c.getUrl(EndpointAddMessage)
	resp, err := c.api.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(endpoint)

	if resp.StatusCode() > 200 || err != nil {
		fmt.Printf("%v: %s\n", resp.StatusCode(), resp.Status())
		return err
	}
	fmt.Println("sent message")
	fmt.Println(resp.StatusCode(), resp.Status())
	fmt.Println(string(resp.Body()))

	err = c.updateConversation(cID, encSnippet, timestamp)
	if err != nil {
		return err
	}

	return nil

}

func (c *Client) Send(data string, conversationID ConversationID) error {

	m := Message{Data: data}
	return c.SendMessage(m, conversationID)

}
