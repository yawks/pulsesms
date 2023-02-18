package pulsesms

import (
	"encoding/json"
	"fmt"
)

type contactID = int

type contact struct {
	ID          contactID `json:"id,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Name        string    `json:"name,omitempty"`
}

func (c *Client) listContacts() ([]contact, error) {

	endpoint := c.getUrl(EndpointContacts)
	max := 100
	offset := 0
	result := []contact{}

	for true {

		resp, err := c.api.R().
			SetQueryParam("account_id", fmt.Sprint(c.accountID)).
			SetQueryParam("limit", fmt.Sprint(max)).
			SetQueryParam("offset", fmt.Sprint(offset)).
			Get(endpoint)

		if err != nil {
			fmt.Println(resp.Status())
			return nil, err

		}

		contacts := []contact{}

		err = json.Unmarshal(resp.Body(), &contacts)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshall contacts: %v", err)
		}

		for _, contact := range contacts {
			err := decryptContact(c.crypto.cipher, &contact)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt contact %v", err)
			}
			result = append(result, contact)

		}
		offset += max
		if len(contacts) < max {
			break
		}
	}

	return result, nil

}
