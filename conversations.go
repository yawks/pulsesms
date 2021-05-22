package pulsesms

import (
	"fmt"
)


type Conversation struct {

}

func (c *Client) List() error {
	index := "index_public_unarchived"

	endpoint := c.getUrl(EndpointConversations)
    fmt.Println("base")
    fmt.Println(endpoint)

    path := fmt.Sprintf("%s/%s", endpoint, index)
	// endpoint := filepath.Join(base, index)


	// result := ]make([]map[string]interface{})

	resp, err := c.api.R().
		SetQueryParam("account_id", c.accountID).
		SetQueryParam("limit", fmt.Sprint(75)).
		Get(path)

	if err != nil {
		fmt.Printf("%v: %s", resp.StatusCode(), resp.Status())
		return err
	}
    fmt.Printf(string(resp.Body()))



    // fmt.Print(len(result))
    // fmt.Print(result)
	// if resp.StatusCode() != 200 {
	//     return fmt.Errorf(resp.Status())
	// }

	// if result.AccountID == "" {
	//     return fmt.Errorf("response missing accounntID")
	// }

	// // TODO decrypt salt
	// c.accountID = result.AccountID

	// fmt.Printf("%+v", result)

	return nil

}
