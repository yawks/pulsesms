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
	ConversationID conversationID `json:"conversation_id,omitempty"`
	DeviceID       DeviceID       `json:"device_id,omitempty"`
	Type           int            `json:"message_type,omitempty"`
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

func (m Message) ChatID() ChatID {
	if m.ConversationID == 0 {
		return fmt.Sprint(m.DeviceID)
	}
	return fmt.Sprint(m.ConversationID)
}

func (m Message) UnixTime() time.Time {
	newTime := m.Timestamp / 1000 >> 0 // remove ms
	return time.Unix(newTime, 0)
}

func (m Message) Received() bool {
	return m.Type == 0 || m.Type == 6
}

func (m Message) Sent() bool {
	return !m.Received()
}

type sendMessageRequest struct {
	AccountID            AccountID      `json:"account_id,omitempty"`
	Data                 string         `json:"data,omitempty"`
	DeviceConversationID conversationID `json:"device_conversation_id,omitempty"`
	DeviceID             DeviceID       `json:"device_id,omitempty"`
	MessageType          int            `json:"message_type,omitempty"`
	MimeType             string         `json:"mime_type,omitempty"`
	Read                 bool           `json:"read,omitempty"`
	Seen                 bool           `json:"seen,omitempty"`
	SentDevice           DeviceID       `json:"sent_device"`
	Timestamp            int64          `json:"timestamp,omitempty"`
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
		SetQueryParam("web", fmt.Sprint(true)).
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

func (c *Client) SendMessage(m Message, chatID string) error {
	convoID, err := strconv.Atoi(chatID)
	if err != nil {
		return fmt.Errorf("invalid chatID")
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

	if m.Timestamp == 0 {
		// js time in ms
		m.Timestamp = time.Now().UTC().UnixNano() / 1e6
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
		DeviceConversationID: convoID,
		DeviceID:             m.ID,
		MessageType:          2,
		Timestamp:            m.Timestamp,
		MimeType:             mimetype,
		Read:                 true,
		Seen:                 true,
		SentDevice:           1,
	}

	endpoint := c.getUrl(EndpointAddMessage)
	resp, err := c.api.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(endpoint)

	fmt.Println(resp.Status())
	if resp.StatusCode() > 200 || err != nil {
		return fmt.Errorf(resp.Status())
	}
	fmt.Println("sent message")

	err = c.updateConversation(convoID, encSnippet, m.Timestamp)
	if err != nil {
		return err
	}

	return nil

}

func (c *Client) Send(data string, chatID ChatID) error {
	m := Message{Data: data}
	return c.SendMessage(m, chatID)

}
